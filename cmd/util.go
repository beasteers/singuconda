package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
)

var OVERLAY_DIR = GetEnvVar("SING_OVERLAY_DIR", "/scratch/work/public/overlay-fs-ext3")
var SIF_DIR = GetEnvVar("SING_SIF_DIR", "/scratch/work/public/singularity")

var DEFAULT_OVERLAY = GetEnvVar("SING_DEFAULT_OVERLAY", "overlay-5GB-200K.ext3.gz")
var DEFAULT_SIF = GetEnvVar("SING_DEFAULT_SIF", "cuda12.3.2-cudnn9.0.0-ubuntu-22.04.4.sif")

const SING_CMD_BLOCK = `singularity exec %s --overlay %s %s /bin/bash << 'EOFXXX'
[[ -e /ext3/env ]] && . /ext3/env > /dev/null
%s
EOFXXX`

const SING_CMD_INTERACTIVE = `
singularity exec %s \
	--overlay %s \
	%s \
	/bin/bash --init-file /ext3/env
`

const SING_CMD_FLEX_SCRIPT = `

# build bash command arguments

# allow commands from stdin (e.g. ./sing <<< "echo hi")
readstdin() {
	read -N1 -t0.5 __  && { (( $? <= 128 )) && { IFS= read -rd '' _stdin; echo "$__$_stdin"; } }
}

# build args
CMD="$(readstdin)"
ARGS=()
if [[ -z "$CMD" ]]; then
	if [[ -z "$SINGUCONDA_NO_INIT_ENV" ]]; then
		ARGS+=(--init-file /ext3/env)
	fi
else
	if [[ -z "$SINGUCONDA_NO_INIT_ENV" ]]; then
		CMD="[[ -e /ext3/env ]] && . /ext3/env;$CMD"
	fi
	ARGS+=(-c "$CMD")
fi

# build singularity arguments

# check for GPUs
GPUS=$(type -p nvidia-smi >&/dev/null && nvidia-smi --query-gpu=name --format=csv,noheader)
NV=$([[ $(echo -n "$GPUS" | grep -v "No devices were found" | awk 'NF' | wc -l) -ge 1 ]] && echo '--nv')
# report GPUs
if [[ -z "$QUIET_SING" ]]; then
	[[ ! -z "$NV" ]] && echo "Detected gpus, using --nv:" && echo $GPUS && echo
	[[ ! -z "$NV" ]] || echo "No gpus detected. Use --nv for gpu bindings." && echo
fi

# get singularity arguments
SCRIPT_DIR="$(dirname ${BASH_SOURCE:-$0})"
SING_NAME="${SING_NAME:-%s}"

OVERLAY="$SCRIPT_DIR/$SING_NAME.ext3"
SIF="$(cat $SCRIPT_DIR/.$SING_NAME.sifpath)"

# run singularity

singularity exec $NV $@ --overlay "${OVERLAY}%s" "$SIF" /bin/bash "${ARGS[@]}"

`

const SINGRW_BLOCK = `
SINGUCONDA_NO_INIT_ENV=1 QUIET_SING=1 ./%srw << 'EOFXXX'
[[ -e /ext3/env ]] && . /ext3/env > /dev/null
%s
EOFXXX
`

// const AUTO_NV = `$([[ $(hostname -s) =~ ^g ]] && echo '--nv')`

const SBATCH_ARG = `#SBATCH --%s=%s\n`

const SSH_LOG_RANDOM_PORT = `
port=$(shuf -i 10000-65500 -n 1)
/usr/bin/ssh -N -f -R $port:localhost:$port log-1
/usr/bin/ssh -N -f -R $port:localhost:$port log-2
/usr/bin/ssh -N -f -R $port:localhost:$port log-3

echo "To access:"
echo "ssh -L $port:localhost:$port $USER@greene.hpc.nyu.edu"
echo "ssh -L $port:localhost:$port greene"
`

func SingCmd(singName string, cmd string) error {
	return RunShell(fmt.Sprintf(SINGRW_BLOCK, singName, cmd))
}

func RunShell(cmd string) error {
	p := exec.Command("bash", "-c", cmd)
	p.Stdin = os.Stdin
	p.Stdout = os.Stdout
	p.Stderr = os.Stderr
	err := p.Run()
	// stdout, err := p.CombinedOutput()
	// fmt.Println(string(stdout))
	return err
}

func StartSing(singName string) error { // overlay string, sif string
	for {
		prompt := promptui.Select{
			Label: "What do you want to do?",
			Items: []string{"nothing, byeee!", "enter (read-only)", "enter (read-write)"},
		}
		// , "install packages", "enter gpu", "enter gpu (write)"
		_, cmd, err := prompt.Run()
		if err != nil {
			return err
		}
		if cmd == "nothing, byeee!" {
			return nil
		}
		if cmd == "enter (read-only)" {
			err = RunShell("./" + singName)
			if err != nil {
				return err
			}
		}
		// if cmd == "enter gpu" {
		// 	err = RunShell("./sing --nv")
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		// if cmd == "install packages" {
		// 	err = InstallPackages(overlay, sif)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		if cmd == "enter (read-write)" {
			err = RunShell("./" + singName + "rw")
			if err != nil {
				return err
			}
		}
		// if cmd == "enter gpu (write)" {
		// 	err = RunShell("./singrw --nv")
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}

}

func WriteSingCmds(singName string, name string) error { //, overlay string, sif string
	script := fmt.Sprintf(SING_CMD_FLEX_SCRIPT, name, ":ro")
	err := os.WriteFile(singName, []byte(script), 0774)
	if err != nil {
		return err
	}

	script = fmt.Sprintf(SING_CMD_FLEX_SCRIPT, name, "")
	err = os.WriteFile(singName+"rw", []byte(script), 0774)
	if err != nil {
		return err
	}
	return nil
}

func HowToRun(singName string, overlay string, sif string) {
	cmd := fmt.Sprintf(SING_CMD_INTERACTIVE, "", overlay+":ro", sif)
	fmt.Printf("To enter the container, run: \033[32m./%s\033[0m \n\nor you can run:\n%s\n", singName, cmd)
	fmt.Printf("The above command opens with read-only. To open with write permissions: \033[32m./%srw\033[0m \n\n", singName)
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return 0
}

func GetEnvVar(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// SortList prioritizes items containing previous choices as substrings
func SortSubstr(allItems, substrings []string) []string {
	// Separate the items based on whether they contain previous choices as substrings
	chosenItems := []string{}
	otherItems := []string{}
	for _, item := range allItems {
		found := false
		for _, choice := range substrings {
			if strings.Contains(item, choice) {
				chosenItems = append(chosenItems, item)
				found = true
				break
			}
		}
		if !found {
			otherItems = append(otherItems, item)
		}
	}

	// Combine the lists with chosen items first
	return append(chosenItems, otherItems...)
}
