package main

import (
	"flag"
	"fmt"
	"github.com/kschjeld/ocputils/pkg/clienthelper"
	v15 "github.com/openshift/client-go/apps/clientset/versioned"
	"github.com/openshift/client-go/apps/informers/externalversions"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	"log"
	"time"
)

func main() {

	flag.Parse()

	config, err := clienthelper.NewOCPClientWithUserconfig()
	if err != nil {
		log.Fatal(err)
	}

	stopchan := make(chan struct{})

	appsclient := v15.NewForConfigOrDie(config)
	appsinformerfactory := externalversions.NewSharedInformerFactory(appsclient, 15 * time.Minute)
	appslister := appsinformerfactory.Apps().V1().DeploymentConfigs().Lister()
	appsinformerfactory.Start(stopchan)
	appsinformerfactory.WaitForCacheSync(stopchan)

	hpaClient := v12.NewForConfigOrDie(config)
	hpas, err := hpaClient.HorizontalPodAutoscalers("").List(v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, hpa := range hpas.Items {
		if hpa.Spec.ScaleTargetRef.Kind != "DeploymentConfig" {
			//log.Printf("%s/%s: References unknown scalable resource %s. Skipping..", hpa.Namespace, hpa.Name, hpa.Spec.ScaleTargetRef.Kind)
			continue
		}

		_, err := appslister.DeploymentConfigs(hpa.Namespace).Get(hpa.Spec.ScaleTargetRef.Name)
		if err != nil {
			fmt.Printf("%s/%s\n", hpa.Namespace, hpa.Name)
			continue
		}
	}
}
