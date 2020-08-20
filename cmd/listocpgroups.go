package main

import (
	"flag"
	"fmt"
	"github.com/kschjeld/ocputils/pkg/clienthelper"
	"github.com/kschjeld/ocputils/pkg/usercache"
	"github.com/openshift/api/user/v1"
	userv1 "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"text/tabwriter"
)

func main() {

	showGroup := flag.String("group", "", "Group to show, or empty to show all")
	searchUserName := flag.String("search", "", "Search for users with given name")
	useSimpleOutput := flag.Bool("simple", false, "Show output using simple Ansible Tower compatible formatting")
	exportDefinitions := flag.String("export", "", "Export definitions into simple text-files in given directory, one pr group")
	exportGroupsFile := flag.String("groupsfile", "", "Create json file with group definitions for use with opt-ansible-groups")
	flag.Parse()

	config, err := clienthelper.NewOCPClientWithUserconfig()
	if err != nil {
		log.Fatal(err)
	}

	userclient, err := userv1.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	userinfo, err := usercache.NewWithClient(userclient)
	if err != nil {
		log.Fatal(err)
	}

	var groupList []v1.Group
	if *showGroup != "" {
		group, err := userclient.Groups().Get(*showGroup, metav1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}
		groupList = append(groupList, *group)
	} else {
		groups, err := userclient.Groups().List(metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		groupList = append(groupList, groups.Items...)
	}

	if *exportDefinitions != "" {
		for _, group := range groupList {
			f, err := os.Create(path.Join(*exportDefinitions, group.Name + ".txt") )
			if err != nil {
				fmt.Printf("Failed to write file: %s\n", err)
				os.Exit(1)
			}
			printGroupSimple(f, group)
			_ = f.Sync()
			if err := f.Close(); err != nil {
				fmt.Printf("Error closing file: %s", err)
			}
		}
		fmt.Printf("Wrote %d group definitions to %s\n", len(groupList), *exportDefinitions)
		return
	}

	if *exportGroupsFile != "" {
		f, err := os.Create(*exportGroupsFile)
		if err != nil {
			fmt.Printf("Failed to open groups file for writing: ", err)
			os.Exit(1)
		}

		err = writeGroupsJSONFile(f, groupList)
		if err != nil {
			fmt.Printf("Error writing groups file: ", err)
			os.Exit(1)
		}

		if err := f.Close(); err != nil {
			fmt.Printf("Error closing file: %s", err)
		}
		return
	}

	if *searchUserName != "" {
		*searchUserName = strings.ToLower(*searchUserName)
		for _, group := range groupList {
			for _, user := range group.Users {
				name := userinfo.GetFullname(user)
				if strings.Contains(strings.ToLower(name), *searchUserName) {
					fmt.Printf("Found: %s (%s) in group %s\n", name, user, group.Name)
				}
			}
		}
		return
	}

	for _, group := range groupList {

		if *useSimpleOutput {
			printGroupSimple(os.Stdout, group)
		} else {
			printGroupFormatted(group, userinfo)
		}
	}
}

func writeGroupsJSONFile(f *os.File, groupslist []v1.Group) error {
	type ocpgroup struct {
		Name string 		`json:"name"`
		Users []string 		`json:"users"`
	}
	var ocpgroups []ocpgroup

	for _, group := range groupslist {
		g := ocpgroup {
			Name: group.Name,
		}
		users := []string{}
		for _, u := range group.Users {
			users = append(users, u)
		}
		sort.Strings(users)
		g.Users = users

		ocpgroups = append(ocpgroups, g)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")

	return enc.Encode(ocpgroups)
}

func printGroupSimple(w io.Writer, group v1.Group) {
	var users []string
	for _, user := range group.Users {
		users = append(users, user)
	}
	sort.Strings(users)

	_, _ = fmt.Fprintf(w, "\n%s\n", group.Name)
	for _, user := range users {
		_, _ = fmt.Fprintln(w, user)
	}
	_, _ = fmt.Fprintf(w, "\n")
}

func printGroupFormatted(group v1.Group, cache *usercache.Usercache) {

	w := new(tabwriter.Writer)
	if group.Users != nil {
		w.Init(os.Stdout, 8, 12, 0, '\t', 0)
		fmt.Printf("Group: %s\n", group.Name)
		fmt.Printf(" Members:\n")
		for _, user := range group.Users {
			fmt.Fprintf(w, " - %s\t%s\n", user, cache.GetFullname(user))
			w.Flush()
		}
	}
	fmt.Println("")
}
