package cmd

import (
	"github.com/manifoldco/promptui"
)

func InstallConda() error {
	err := SingCmd(`
# download miniconda
CONDAURL="https://repo.continuum.io/miniconda/Miniconda3-latest-Linux-x86_64.sh"
CONDASH="Miniconda3-latest-Linux-x86_64.sh"
CONDADIR="/ext3/miniconda3"
echo Miniconda Install Location: $CONDADIR
if [ ! -e "$CONDADIR" ] && [ ! -z $CONDADIR ]; then
	echo installing miniconda inside container...
	echo URL: $CONDAURL
	echo Script Location: $CONDASH
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

	err = SingCmd(`
# write environment file

cat > /ext3/env << 'EOFENV'
#!/bin/bash
export PATH=/ext3/miniconda3/bin:$PATH
source /ext3/miniconda3/etc/profile.d/conda.sh -y
[[ -f /ext3/conda.activate ]] && source /ext3/conda.activate

if [[ -z "$QUIET_SING" ]]; then
echo "hello :) your python:" "$(type -P python)"
python --version 2>&1
fi

EOFENV
chmod +x /ext3/env
	`)
	if err != nil {
		return err
	}

	err = SingCmd(`
# show conda/python info
conda info --envs
type -P python
echo "You're currently setup with:"
python --version
	`)
	if err != nil {
		return err
	}

	promptyn := promptui.Prompt{
		Label:   "Want a different python version? (e.g. 3.8, 3.6) If no, leave blank and press enter. To use the base environment, use \"-\"",
		Default: "",
	}
	pythonVersion, err := promptyn.Run()
	if err != nil {
		return err
	}
	err = SingCmd(`
PYVER="` + pythonVersion + `"

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

	err = SingCmd(`
	echo Updating conda and pip...
	conda update -n base conda -yq
	conda install pip -yq
	`)
	if err != nil {
		return err
	}
	return nil
}

// func InstallPackages(overlay string, sif string) error {
// 	for {
// 		prompt := promptui.Select{
// 			Label: "Do you want to install any packages?",
// 			Items: []string{"nope I'm good!", "conda install", "pip install", "pip install -r", "pip install -e"},
// 		}
// 		_, cmd, err := prompt.Run()
// 		if err != nil {
// 			return err
// 		}
// 		if cmd == "nope I'm good!" {
// 			return nil
// 		}

// 		prompt1 := promptui.Prompt{
// 			Label: cmd,
// 		}
// 		installs, err := prompt1.Run()
// 		if err != nil {
// 			return err
// 		}
// 		if installs != "" {
// 			SingCmd(fmt.Sprintf("%s %s", cmd, installs))
// 		}
// 	}
// }
