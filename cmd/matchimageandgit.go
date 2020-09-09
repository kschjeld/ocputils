package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/kschjeld/ocputils/pkg/clienthelper"
	v12 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	imageV1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log"
	"sort"
	"strings"
	simplejson "github.com/bitly/go-simplejson"
)

/*
loop oc projects ($project)
oc project $project
loop oc get is ($is)
 oc describe is $is | grep '* docker-registry'
oc describe image sha256:de162d9ca1aadb31c0e9be9a0639a050aa2dc69694e6d797eaa81c5eec0425f2 | grep no.telenor.git.url
no.telenor.git.url=ssh://git@prima.corp.telenor.no:7999/~t940807/t940807-the-best-ever-webservice-star.git   – extract Git project
 */

const Label_GitUrl = "git.url"


func main() {

	flag.Parse()

	config, err := clienthelper.NewOCPClientWithUserconfig()
	if err != nil {
		log.Fatal(err)
	}

	namespaceClient, err := coreV1.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	imageClient,err := imageV1.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	appsClient, err := v12.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	nsList, err := namespaceClient.Namespaces().List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	var unmappedImages []string

	//Header
	fmt.Println(" OS NameSpace \t OS Deployment Config \t OS Image \t BitBucket Project \t Bit Bucket Repo")


	for _, ns := range nsList.Items {
		nsName := ns.Name

		dcs, err := appsClient.DeploymentConfigs(nsName).List(metav1.ListOptions{})
		if err != nil {
			panic(err)
		}

		for _, dc := range dcs.Items {

			for _, c := range dc.Spec.Template.Spec.Containers {

				// Skip sidecar containers
				if strings.Contains(c.Name, "sidecar") {
					continue
				}

				if !strings.Contains(c.Image, "@") {
					// Will only look at images referenced with SHA
					continue
				}
				image, err := imageClient.Images().Get(getImageSha(c.Image), metav1.GetOptions{})
				if err != nil {
					fmt.Printf("Error getting Image: %s\n", err)
					continue
				}

				gitUrl, err := extractGitUrl(image.DockerImageMetadata.Raw)
				if err != nil {
					fmt.Printf("Error extracting git url: %s\n", err)
					continue
				}

				if gitUrl != "" {
					parts := strings.Split(gitUrl, "/")
					if len(parts) != 5 {
						fmt.Printf("Error. Git URL should have had 5 parts" +  gitUrl)
					}
					fmt.Printf("%s \t %s \t %s \t %s \t %s \n", nsName, dc.Name ,c.Image, parts[3],parts[4])
				} else {
					unmappedImages = append(unmappedImages, nsName + "\t" + dc.Name + "\t" + c.Image)
				}
			}
		}
	}

	sort.Strings(unmappedImages)
	for _, i := range unmappedImages {
		fmt.Println( i + " \t \t ")
	}
}

func extractGitUrl(raw []byte) (string, error) {
	json, err := simplejson.NewFromReader(bytes.NewReader(raw))
	if err != nil {
		return "", err
	}

	if json != nil {
		labels := json.GetPath("Config", "Labels")
		if labels != nil {
			labelsMap, err := labels.Map()
			if labelsMap == nil || err != nil {
				return "", nil
			}
			for l, v := range labelsMap {
				// We have labelled with whitespace in front of key, so can not simply look up by relevant key
				if strings.Contains(l, Label_GitUrl) {
					return v.(string), nil
				}
			}
		}
	}

	return "", nil
}

func getImageSha(image string) string {
	s := strings.Split(image, "@")
	return s[len(s)-1]
}
