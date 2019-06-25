// Copyright 2016-2019, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/pulumi/kubespy/watch"
	"github.com/spf13/cobra"
	"github.com/yudai/gojsondiff"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiwatch "k8s.io/apimachinery/pkg/watch"
)

func init() {
	rootCmd.AddCommand(recordCmd)
}

var recordCmd = &cobra.Command{
	Use:   "record <apiVersion> <kind> [<namespace>/]<name>",
	Short: "Displays events generated by a Kubernetes resource in real time. Emitted as a JSON array.",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		namespace, name, err := parseObjID(args[2])
		if err != nil {
			log.Fatal(err)
		}

		events, err := watch.Forever(args[0], args[1], watch.ThisObject(namespace, name))
		if err != nil {
			log.Fatal(err)
		}

		var last *unstructured.Unstructured
		for {
			select {
			case e := <-events:
				o := e.Object.(*unstructured.Unstructured)
				switch e.Type {
				case apiwatch.Added:
					fmt.Println("[")
					fmt.Print("  ")
					if output, err := json.MarshalIndent(o.Object, "  ", "  "); err != nil {
						log.Fatal(err)
					} else {
						fmt.Print(string(output))
					}
				case apiwatch.Modified:
					diff := gojsondiff.New().CompareObjects(last.Object, o.Object)
					if diff.Modified() {
						fmt.Println(",")
						fmt.Print("  ")
						if output, err := json.MarshalIndent(o.Object, "  ", "  "); err != nil {
							log.Fatal(err)
						} else {
							fmt.Print(string(output))
						}
					}
				case apiwatch.Deleted:
					fmt.Println("\n]")
					os.Exit(0)
				}
				last = o
			}
		}
	},
}
