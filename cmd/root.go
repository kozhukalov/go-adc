/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package cmd

import (
	"github.com/spf13/cobra"
	"io"
	"jinr.ru/greenlab/go-adc/cmd/config"
	"jinr.ru/greenlab/go-adc/cmd/control"
	"jinr.ru/greenlab/go-adc/cmd/discover"
	"jinr.ru/greenlab/go-adc/cmd/mstream"
	"jinr.ru/greenlab/go-adc/cmd/completion"
	"jinr.ru/greenlab/go-adc/pkg/log"
)

func NewRootCommand(out io.Writer) (*cobra.Command) {
	cmd := &cobra.Command{
		Use:           "go-adc",
		Short:         "Tool to work with ADC64 devices",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Init(cmd.ErrOrStderr())
		},
	}
	cmd.SetOut(out)
	cmd.AddCommand(config.NewCommand())
	cmd.AddCommand(control.NewCommand())
	cmd.AddCommand(discover.NewCommand())
	cmd.AddCommand(mstream.NewCommand())
	cmd.AddCommand(completion.NewCommand())
	return cmd
}
