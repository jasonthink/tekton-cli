// Copyright © 2019 The Tekton Authors.
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

package taskrun

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/formatted"
	traction "github.com/tektoncd/cli/pkg/taskrun"
	"github.com/tektoncd/cli/pkg/validate"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	succeeded        = formatted.ColorStatus("Succeeded")
	failed           = formatted.ColorStatus("Failed")
	taskrunCancelled = formatted.ColorStatus("Cancelled") + "(TaskRunCancelled)"
)

func cancelCommand(p cli.Params) *cobra.Command {
	eg := `Cancel the TaskRun named 'foo' from namespace 'bar':

    tkn taskrun cancel foo -n bar
`

	c := &cobra.Command{
		Use:          "cancel",
		Short:        "Cancel a TaskRun in a namespace",
		Example:      eg,
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		Annotations: map[string]string{
			"commandType": "main",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			s := &cli.Stream{
				Out: cmd.OutOrStdout(),
				Err: cmd.OutOrStderr(),
			}

			if err := validate.NamespaceExists(p); err != nil {
				return err
			}

			return cancelTaskRun(p, s, args[0])
		},
	}

	_ = c.MarkZshCompPositionalArgumentCustom(1, "__tkn_get_taskrun")
	return c
}

func cancelTaskRun(p cli.Params, s *cli.Stream, trName string) error {
	cs, err := p.Clients()
	if err != nil {
		return fmt.Errorf("failed to create tekton client")
	}

	tr, err := traction.Get(cs, trName, metav1.GetOptions{}, p.Namespace())
	if err != nil {
		return fmt.Errorf("failed to find taskrun: %s", trName)
	}

	taskrunCond := formatted.Condition(tr.Status.Conditions)
	if taskrunCond == succeeded || taskrunCond == failed || taskrunCond == taskrunCancelled {
		return fmt.Errorf("failed to cancel taskrun %s: taskrun has already finished execution", trName)
	}

	tr.Spec.Status = v1beta1.TaskRunSpecStatusCancelled
	if _, err := traction.Update(cs, tr, metav1.UpdateOptions{}, p.Namespace()); err != nil {
		return fmt.Errorf("failed to cancel taskrun %q: %s", trName, err)
	}

	fmt.Fprintf(s.Out, "TaskRun cancelled: %s\n", tr.Name)
	return nil
}
