package gslb

import (
	"context"
	"fmt"
	"os"

	k8gbv1beta1 "github.com/AbsaOSS/k8gb/pkg/apis/k8gb/v1beta1"
)

func (r *ReconcileGslb) finalizeGslb(gslb *k8gbv1beta1.Gslb) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.

	gslbZoneName := os.Getenv("DNS_ZONE")
	if len(os.Getenv("INFOBLOX_GRID_HOST")) > 0 {
		objMgr, err := infobloxConnection()
		if err != nil {
			return err
		}
		findZone, err := objMgr.GetZoneDelegated(gslbZoneName)
		if err != nil {
			return err
		}

		if findZone != nil {
			err = checkZoneDelegated(findZone, gslbZoneName)
			if err != nil {
				return err
			}
			if len(findZone.Ref) > 0 {
				log.Info(fmt.Sprintf("Deleting delegated zone(%s)...", gslbZoneName))
				_, err := objMgr.DeleteZoneDelegated(findZone.Ref)
				if err != nil {
					return err
				}
			}
		}

		edgeDNSZone := os.Getenv("EDGE_DNS_ZONE")
		clusterGeoTag := os.Getenv("CLUSTER_GEO_TAG")
		heartbeatTXTName := fmt.Sprintf("%s-heartbeat-%s.%s", gslb.Name, clusterGeoTag, edgeDNSZone)
		findTXT, err := objMgr.GetTXTRecord(heartbeatTXTName)
		if err != nil {
			return err
		}

		if findTXT != nil {
			if len(findTXT.Ref) > 0 {
				log.Info(fmt.Sprintf("Deleting split brain TXT record(%s)...", heartbeatTXTName))
				_, err := objMgr.DeleteTXTRecord(findTXT.Ref)
				if err != nil {
					return err
				}

			}
		}
	}

	log.Info("Successfully finalized Gslb")
	return nil
}

func (r *ReconcileGslb) addFinalizer(gslb *k8gbv1beta1.Gslb) error {
	log.Info("Adding Finalizer for the Gslb")
	gslb.SetFinalizers(append(gslb.GetFinalizers(), gslbFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), gslb)
	if err != nil {
		log.Error(err, "Failed to update Gslb with finalizer")
		return err
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}
