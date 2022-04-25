// Copyright 2019-2022 The Inspektor Gadget authors
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

package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kinvolk/inspektor-gadget/cmd/kubectl-gadget/utils"
	"github.com/kinvolk/inspektor-gadget/pkg/gadgets/bindsnoop/types"
	eventtypes "github.com/kinvolk/inspektor-gadget/pkg/types"
	"github.com/spf13/cobra"
)

var (
	targetPid    uint
	targetPorts  []uint
	ignoreErrors bool
)

var bindsnoopCmd = &cobra.Command{
	Use:   "bind",
	Short: "Trace the kernel functions performing socket binding",
	RunE: func(cmd *cobra.Command, args []string) error {
		// print header
		switch params.OutputMode {
		case utils.OutputModeCustomColumns:
			fmt.Println(getCustomBindsnoopColsHeader(params.CustomColumns))
		case utils.OutputModeColumns:
			fmt.Printf("%-16s %-16s %-16s %-16s %-6s %-16s %-6s %-16s %-6s %-6s %s\n",
				"NODE", "NAMESPACE", "POD", "CONTAINER",
				"PID", "COMM", "PROTO", "ADDR", "PORT", "OPTS", "IF")
		}

		portsStringSlice := []string{}
		for _, port := range targetPorts {
			portsStringSlice = append(portsStringSlice, strconv.FormatUint(uint64(port), 10))
		}

		config := &utils.TraceConfig{
			GadgetName:       "bindsnoop",
			Operation:        "start",
			TraceOutputMode:  "Stream",
			TraceOutputState: "Started",
			CommonFlags:      &params,
			Parameters: map[string]string{
				"pid":           strconv.FormatUint(uint64(targetPid), 10),
				"ports":         strings.Join(portsStringSlice, ","),
				"ignore_errors": strconv.FormatBool(ignoreErrors),
			},
		}

		err := utils.RunTraceAndPrintStream(config, bindsnoopTransformLine)
		if err != nil {
			return utils.WrapInErrRunGadget(err)
		}

		return nil
	},
}

func init() {
	TraceCmd.AddCommand(bindsnoopCmd)
	utils.AddCommonFlags(bindsnoopCmd, &params)

	bindsnoopCmd.PersistentFlags().UintVarP(
		&targetPid,
		"pid",
		"",
		0,
		"Show only bind events generated by this particular PID",
	)
	bindsnoopCmd.PersistentFlags().UintSliceVarP(
		&targetPorts,
		"ports",
		"P",
		[]uint{},
		"Trace only bind events involving these ports",
	)
	bindsnoopCmd.PersistentFlags().BoolVarP(
		&ignoreErrors,
		"ignore-errors",
		"i",
		true,
		"Show only events where the bind succeeded",
	)
}

// bindsnoopTransformLine is called to transform an event to columns
// format according to the parameters
func bindsnoopTransformLine(line string) string {
	var sb strings.Builder
	var e types.Event

	if err := json.Unmarshal([]byte(line), &e); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", utils.WrapInErrUnmarshalOutput(err, line))
		return ""
	}

	if e.Type == eventtypes.ERR || e.Type == eventtypes.WARN ||
		e.Type == eventtypes.DEBUG || e.Type == eventtypes.INFO {
		fmt.Fprintf(os.Stderr, "%s: node %s: %s", e.Type, e.Node, e.Message)
		return ""
	}

	if e.Type != eventtypes.NORMAL {
		return ""
	}
	switch params.OutputMode {
	case utils.OutputModeColumns:
		sb.WriteString(fmt.Sprintf("%-16s %-16s %-16s %-16s %-6d %-16s %-6s %-16s %-6d %-6s %s",
			e.Node, e.Namespace, e.Pod, e.Container,
			e.Pid, e.Comm, e.Protocol, e.Addr, e.Port, e.Options, e.Interface))
	case utils.OutputModeCustomColumns:
		for _, col := range params.CustomColumns {
			switch col {
			case "node":
				sb.WriteString(fmt.Sprintf("%-16s", e.Node))
			case "namespace":
				sb.WriteString(fmt.Sprintf("%-16s", e.Namespace))
			case "pod":
				sb.WriteString(fmt.Sprintf("%-16s", e.Pod))
			case "container":
				sb.WriteString(fmt.Sprintf("%-16s", e.Container))
			case "pid":
				sb.WriteString(fmt.Sprintf("%-6d", e.Pid))
			case "comm":
				sb.WriteString(fmt.Sprintf("%-16s", e.Comm))
			case "proto":
				sb.WriteString(fmt.Sprintf("%-6s", e.Protocol))
			case "addr":
				sb.WriteString(fmt.Sprintf("%-16s", e.Addr))
			case "port":
				sb.WriteString(fmt.Sprintf("%-6d", e.Port))
			case "opts":
				sb.WriteString(fmt.Sprintf("%-6s", e.Options))
			case "if":
				sb.WriteString(fmt.Sprintf("%-6s", e.Interface))
			}
			sb.WriteRune(' ')
		}
	}

	return sb.String()
}

func getCustomBindsnoopColsHeader(cols []string) string {
	var sb strings.Builder

	for _, col := range cols {
		switch col {
		case "node":
			sb.WriteString(fmt.Sprintf("%-16s", "NODE"))
		case "namespace":
			sb.WriteString(fmt.Sprintf("%-16s", "NAMESPACE"))
		case "pod":
			sb.WriteString(fmt.Sprintf("%-16s", "POD"))
		case "container":
			sb.WriteString(fmt.Sprintf("%-16s", "CONTAINER"))
		case "pid":
			sb.WriteString(fmt.Sprintf("%-6s", "PID"))
		case "comm":
			sb.WriteString(fmt.Sprintf("%-16s", "COMM"))
		case "proto":
			sb.WriteString(fmt.Sprintf("%-6s", "PROTO"))
		case "addr":
			sb.WriteString(fmt.Sprintf("%-16s", "ADDR"))
		case "port":
			sb.WriteString(fmt.Sprintf("%-6s", "PORT"))
		case "opts":
			sb.WriteString(fmt.Sprintf("%-6s", "OPTS"))
		case "if":
			sb.WriteString(fmt.Sprintf("%-6s", "IF"))
		}
		sb.WriteRune(' ')
	}

	return sb.String()
}
