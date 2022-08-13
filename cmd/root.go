/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const OVERLAY_DIR = "/scratch/work/public/overlay-fs-ext3"
const SIF_DIR = "/scratch/work/public/singularity"

const DEFAULT_OVERLAY = "overlay-5GB-200K.ext3.gz"
const DEFAULT_SIF = "cuda11.0-cudnn8-devel-ubuntu18.04.sif"

const SING_CMD_BLOCK = `singularity exec %s --overlay %s %s /bin/bash << 'EOFXXX'
[[ -e /ext3/env ]] && . /ext3/env > /dev/null
%s
EOFXXX`

const SING_CMD_INTERACTIVE = "singularity exec %s \\\n\t--overlay %s \\\n\t%s \\\n\t/bin/bash --init-file /ext3/env"

const AUTO_NV = `$([[ $(hostname -s) =~ ^g ]] && echo '--nv')`

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

func singCmd(overlay string, sif string, cmd string) error {
	return runShell(fmt.Sprintf(SING_CMD_BLOCK, "", overlay, sif, cmd))
}

func runShell(cmd string) error {
	p := exec.Command("bash", "-c", cmd)
	p.Stdin = os.Stdin
	p.Stdout = os.Stdout
	p.Stderr = os.Stderr
	err := p.Run()
	// stdout, err := p.CombinedOutput()
	// fmt.Println(string(stdout))
	return err
}

func getOverlay() (string, string, error) {
	// look for existing overlays in this directory
	existingMatches, err := filepath.Glob("*.ext3")
	if err != nil {
		return "", "", err
	}

	// select from existing overlays
	if len(existingMatches) > 0 {
		prompt1 := promptui.Select{
			Label: "There are overlays in this directory. Use one?",
			Items: append(existingMatches, "new..."),
		}
		_, existingOverlay, err := prompt1.Run()
		if err != nil {
			return "", "", err
		}
		if existingOverlay != "new..." {
			overlayName := strings.TrimSuffix(path.Base(existingOverlay), ".gz")
			existingOverlay, _ = filepath.Abs(existingOverlay)
			return existingOverlay, overlayName, nil
		}
	}

	// select new overlay
	matches, err := filepath.Glob(filepath.Join(OVERLAY_DIR, "*.ext3.gz"))
	if err != nil {
		return "", "", err
	}

	searcher := func(input string, index int) bool {
		name := strings.Replace(strings.ToLower(matches[index]), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt2 := promptui.Select{
		Label:             "Which overlay to use?",
		Items:             matches,
		Searcher:          searcher,
		StartInSearchMode: true,
		CursorPos:         indexOf(filepath.Join(OVERLAY_DIR, DEFAULT_OVERLAY), matches),
	}
	_, overlayPath, err := prompt2.Run()
	if err != nil {
		return "", "", err
	}

	// give the overlay a new name
	defaultOverlayName := path.Base(overlayPath)
	defaultOverlayName = strings.TrimSuffix(defaultOverlayName, ".gz")
	defaultOverlayName = strings.TrimSuffix(defaultOverlayName, filepath.Ext(defaultOverlayName))
	prompt3 := promptui.Prompt{
		Label:   "Why don't you give your overlay a name?",
		Default: defaultOverlayName,
	}
	name, err := prompt3.Run()
	if err != nil {
		return "", "", err
	}
	if name == "" {
		name = defaultOverlayName
	}
	fmt.Printf("You choose %q\n", name)

	overlayDest := fmt.Sprintf("%s.ext3", name)
	if _, err := os.Stat(overlayDest); !os.IsNotExist(err) {
		fmt.Printf("file exists %s\n", overlayDest)
		return "", "", err
	}

	// expand the overlay to the current directory
	fmt.Printf("Unzipping %s to %s...\n", overlayPath, overlayDest)
	f, err := os.Open(overlayPath)
	if err != nil {
		return "", "", err
	}
	reader, err := gzip.NewReader(f)
	if err != nil {
		return "", "", err
	}
	defer reader.Close()

	o, err := os.Create(overlayDest)
	if err != nil {
		return "", "", err
	}
	defer o.Close()
	_, err = o.ReadFrom(reader)
	if err != nil {
		return "", "", err
	}
	fmt.Printf("Done!\n")
	return overlayDest, name, nil
}

func getSif(name string) (string, error) {
	sifCache := fmt.Sprintf(".%s.sifpath", name)

	// check if we configured the sif file before
	defaultSif := filepath.Join(SIF_DIR, DEFAULT_SIF)
	if _, err := os.Stat(sifCache); err == nil {
		buf, err := os.ReadFile(sifCache)
		if err != nil {
			return "", err
		}
		defaultSif = string(buf)

		promptyn := promptui.Prompt{
			Label:     fmt.Sprintf("Use %s", defaultSif),
			IsConfirm: true,
			Default:   "y",
		}
		_, err = promptyn.Run()
		if err == nil {
			return defaultSif, nil
		}
	}

	// select from sifs
	matches, err := filepath.Glob(filepath.Join(SIF_DIR, "*.sif"))
	if err != nil {
		return "", err
	}

	searcher := func(input string, index int) bool {
		name := strings.Replace(strings.ToLower(matches[index]), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}
	prompt := promptui.Select{
		Label:             "Which sif to use?",
		Items:             matches,
		Searcher:          searcher,
		StartInSearchMode: true,
		CursorPos:         indexOf(defaultSif, matches),
	}
	_, sifPath, err := prompt.Run()
	if err != nil {
		return "", err
	}

	// write cache for next time
	err = os.WriteFile(sifCache, []byte(sifPath), 0774)
	if err != nil {
		return "", err
	}

	return sifPath, nil
}

func installConda(overlay string, sif string) error {

	err := singCmd(overlay, sif, `
	# download miniconda

	CONDAURL=https://repo.continuum.io/miniconda/Miniconda3-latest-Linux-x86_64.sh
	CONDASH=Miniconda3-latest-Linux-x86_64.sh
	CONDADIR=/ext3/miniconda3
	if [ ! -e "$CONDADIR" ]; then
		echo installing miniconda inside container...
		[[ ! -f "$CONDASH" ]] && wget "$CONDAURL"
		bash "$CONDASH" -b -p "$CONDADIR"
		rm "$CONDASH"
		echo "================================="
		echo "Installed miniconda"
		echo 
	else
		echo miniconda exists: "$CONDADIR"
	fi
	`)
	if err != nil {
		return err
	}

	err = singCmd(overlay, sif, `
# write environment file

cat > /ext3/env << 'EOFENV'
#!/bin/bash
export PATH=/ext3/miniconda3/bin:$PATH
source /ext3/miniconda3/etc/profile.d/conda.sh -y
[[ -f /ext3/conda.activate ]] && source /ext3/conda.activate
echo "hello :) you're using:" "$(which python)"
python --version 2>&1
EOFENV
chmod +x /ext3/env
	`)
	if err != nil {
		return err
	}

	err = singCmd(overlay, sif, `
	# show conda/python info
	conda info --envs
	which python
	echo "You're currently setup with:"
	python --version
	`)
	if err != nil {
		return err
	}

	promptyn := promptui.Prompt{
		Label:   "Want a different python version? (e.g. 3.8, 3.6) If no, leave blank. To use the base environment, use \"-\"",
		Default: "",
	}
	pythonVersion, err := promptyn.Run()
	if err != nil {
		return err
	}
	err = singCmd(overlay, sif, `
	PYVER="`+pythonVersion+`"

	# python version special cases

	if [[ -z "$PYVER" ]]; then
		echo "keeping environment..."
		exit 0
	fi

	if [[ "$PYVER" == "-" ]]; then
		echo "resetting to the base environment..."
		echo "" > /ext3/conda.activate
		exit 0
	fi

	# install the environment if it doesnt already exist

	if [[ -d /ext3/miniconda3/envs/$PYVER ]]; then
		CONDAENV=$PYVER
		echo "using environment: $CONDAENV"
	else
		CONDAENV="py${PYVER//[^0-9]/}"
		echo "using environment: $CONDAENV"
		if [[ ! -d /ext3/miniconda3/envs/$CONDAENV ]]; then
			echo "creating environment: $CONDAENV"
			export PATH=/ext3/miniconda3/bin:$PATH
			conda create -n "$CONDAENV" python="$PYVER"
		fi
	fi

	# add script to activate this environment

	if [[ ! -z "$CONDAENV" ]]; then
		echo "conda activate $CONDAENV" > /ext3/conda.activate
	fi
	`)
	if err != nil {
		return err
	}

	err = singCmd(overlay, sif, `
	conda update -n base conda -yq
	conda install pip -yq
	`)
	if err != nil {
		return err
	}
	return nil
}

func installPackages(overlay string, sif string) error {
	for {
		prompt := promptui.Select{
			Label: "Do you want to install any packages?",
			Items: []string{"nope I'm good!", "conda install", "pip install", "pip install -r", "pip install -e"},
		}
		_, cmd, err := prompt.Run()
		if err != nil {
			return err
		}
		if cmd == "nope I'm good!" {
			return nil
		}

		prompt1 := promptui.Prompt{
			Label: cmd,
		}
		installs, err := prompt1.Run()
		if err != nil {
			return err
		}
		if installs != "" {
			singCmd(overlay, sif, fmt.Sprintf("%s %s", cmd, installs))
		}
	}
}

func startSing(overlay string, sif string) error {
	for {
		prompt := promptui.Select{
			Label: "What do you want to do?",
			Items: []string{"nothing, byeee!", "enter", "enter gpu", "install packages", "enter (write)", "enter gpu (write)"},
		}
		_, cmd, err := prompt.Run()
		if err != nil {
			return err
		}
		if cmd == "nothing, byeee!" {
			return nil
		}
		if cmd == "enter" {
			err = runShell("./sing")
			if err != nil {
				return err
			}
		}
		if cmd == "enter gpu" {
			err = runShell("./sing --nv")
			if err != nil {
				return err
			}
		}
		if cmd == "install packages" {
			err = installPackages(overlay, sif)
			if err != nil {
				return err
			}
		}
		if cmd == "enter (write)" {
			err = runShell("./singrw")
			if err != nil {
				return err
			}
		}
		if cmd == "enter gpu (write)" {
			err = runShell("./singrw --nv")
			if err != nil {
				return err
			}
		}
	}

}

func writeSingCmds(overlay string, sif string) error {
	overlay, _ = filepath.Abs(overlay)

	cmd := fmt.Sprintf(SING_CMD_INTERACTIVE, "$@", overlay+":ro", sif)
	fmt.Printf("To enter the container, run: \033[32m./sing\033[0m \n\nwhich is equivalent to:\n%s\n", cmd)
	err := os.WriteFile("sing", []byte(cmd), 0774)
	if err != nil {
		return err
	}

	fmt.Printf("\nTo use GPUs do: \033[32m./sing --nv\033[0m\n")
	cmd = fmt.Sprintf(SING_CMD_INTERACTIVE, "$@", overlay, sif)
	fmt.Printf("The above command opens with read-only. To open with write permissions: \033[32m./singrw\033[0m \n\n")
	err = os.WriteFile("singrw", []byte(cmd), 0774)
	if err != nil {
		return err
	}
	return nil
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return 0
}

// #!/bin/bash
// #SBATCH --job-name=jlb
// #SBATCH --nodes=1
// #SBATCH --gres=gpu:0
// #SBATCH --cpus-per-task=1
// #SBATCH --mem=32GB
// #SBATCH --time=7:00:00
// #SBATCH --mail-user=bs3639@nyu.edu
// #SBATCH --output="/home/bs3639/logs/jupyter.out"

// func jupyterSing(overlay string, sif string) error {
// 	cmd := fmt.Sprintf(
// 		fmt.Sprintf(SBATCH_ARG, "job-name",  "jupyter") +
// 		fmt.Sprintf(SBATCH_ARG, "output",  "jupyter.out") +
// 		SSH_LOG_RANDOM_PORT + SING_CMD_BLOCK, '', overlay, sif, `
// 	python -m ipykernel install --name ext-miniconda --user
// 	jupyter lab --no-browser --port $port --notebook-dir=$(pwd);
// 	`);

// 	err := os.WriteFile("jupyter", []byte(cmd), 0774)
// }

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "singuconda",
	Short: "Let's build a singularity+anaconda container <3 asdsafjaskd !!!",
	Long: `Let's build a singularity+anaconda container <3 asdsafjaskd !!!

First, make sure you are cd'd in the directory where you want your overlay to live.

Then run:
	~/singuconda

What happens:
	1. pick an overlay file
	2. pick a sif file
	3. install miniconda and optionally a specific python version
	4. adds startup environment script (/ext3/env)
	5. menu to install packages in the container
	6. create shortcut script(s) for running the container

Then you're all done!

You can re-run it if you want to change anything (sif file, python version, installs).

This was built for NYU Greene's environment, but it should apply elsewhere too!`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// configure overlay/sif
		overlay, name, err := getOverlay()
		if err != nil {
			fmt.Println(err)
			return
		}
		sif, err := getSif(name)
		if err != nil {
			fmt.Println(err)
			return
		}

		// download and install miniconda (if not done already)
		err = installConda(overlay, sif)
		if err != nil {
			fmt.Println(err)
			return
		}

		// we're all good! write out shortcuts
		fmt.Printf("\nGreat you're all set!\n\n")
		err = writeSingCmds(overlay, sif)
		if err != nil {
			fmt.Println(err)
			return
		}

		// provide quick actions to get started
		err = startSing(overlay, sif)
		fmt.Printf("\nHappy training! :)\n")
		fmt.Printf("\nQuick commands: \033[32m./sing\033[0m    \033[32m./sing --nv\033[0m    \033[32m./singrw\033[0m \n")
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.singuconda.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
