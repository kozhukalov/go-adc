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

package reg

import (
	"github.com/spf13/cobra"
)

const (
	DeviceOptionName = "device"
	AddrOptionName = "addr"
	ValueOptionName = "value"
)

func NewRegCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "reg",
		Short:         "Low level control by means of reading from/writing to registers",
	}
	cmd.AddCommand(NewReadCommand())
	cmd.AddCommand(NewWriteCommand())
	return cmd
}
