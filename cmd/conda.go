package cmd

import (
	"github.com/manifoldco/promptui"
)

func InstallConda(singName string) error {
	err := SingCmd(singName, `
# download miniconda
CONDAURL="https://repo.continuum.io/miniconda/Miniconda3-latest-Linux-x86_64.sh"
CONDASH="Miniconda3-latest-Linux-x86_64.sh"
CONDADIR="/ext3/miniconda3"
echo Miniconda Install Location: $CONDADIR
if [ ! -e "$CONDADIR" ] && [ ! -z $CONDADIR ]; then
	echo installing miniconda inside container...
	echo URL: $CONDAURL
	echo Script Location: $CONDASH
	[[ ! -f "$CONDASH" ]] && wget --no-check-certificate "$CONDAURL"
	bash "$CONDASH" -b -p "$CONDADIR"
	rm "$CONDASH"
	"$CONDADIR"/condabin/conda tos accept
	"$CONDADIR"/condabin/conda update -n base conda -yq || echo "Couldn't update conda"
	"$CONDADIR"/condabin/conda clean -yqa || echo "Couldn't clean conda"
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

	err = SingCmd(singName, `
# write environment file

cat > /ext3/env << 'EOFENV'
#!/bin/bash
export CONDA_PLUGINS_AUTO_ACCEPT_TOS="true"
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

	for {
		err = InstallEnv(singName)
		if err == nil {
			return nil
		}
	}
}
func InstallEnv(singName string) error {
	err := SingCmd(singName, `
# show conda/python info
conda env list
type -P python
echo "You're currently setup with:"
python --version

echo "To keep this version: Leave blank."
# [[ ! -z $(cat /ext3/conda.activate) ]] && echo "To revert to the base environment, use \"-\""
	`)
	if err != nil {
		return err
	}
	promptyn := promptui.Prompt{
		Label:   "Want a different python version? (e.g. 3.10, 3.12.1)",
		Default: "",
	}
	pythonVersion, err := promptyn.Run()
	if err != nil {
		return err
	}
	err = SingCmd(singName, `
PYVER="`+pythonVersion+`"

# python version special cases

# they left it blank
if [[ -z "$PYVER" ]]; then
	echo "keeping environment..."
	exit 0
fi

# revert to the base environment. honestly, you could just type "base" instead, so may delete.
if [[ "$PYVER" == "-" ]] || [[ "$PYVER" == "base" ]]; then
	echo "resetting to the base environment..."
	echo "" > /ext3/conda.activate
	exit 0
fi

# install the environment if it doesnt already exist

# first check if the user typed a literal environment
if [[ -d /ext3/miniconda3/envs/$PYVER ]]; then
	CONDAENV=$PYVER
	echo "using environment: $CONDAENV"
else
	# otherwise lets try and parse out a sensible environment name

	# parse out the version number
	VERSION=""
	if [[ $PYVER =~ ^([a-zA-Z]*)([0-9.]*)([a-zA-Z]*)$ ]]; then
		# Example 1: PYVER=3.11 -> NAME=py311 VERSION=3.11
		# Example 2: PYVER=3.11.2 -> NAME=py311_2 VERSION=3.11.2
		# Example 3: PYVER=3.11asdf -> NAME=py311asdf VERSION=3.11
		# Example 4: PYVER=blah3.11 -> NAME=blah VERSION=3.11
		PRE="${BASH_REMATCH[1]}"
		VERSION="${BASH_REMATCH[2]}"
		POST="${BASH_REMATCH[3]}"
		VER=${VERSION/./}
		VER=${VER//./_}
		CONDAENV="${PRE:-py$VER$POST}"
		echo "parsed environment name: $CONDAENV"
		echo "parsed version: $VERSION"
	fi

	# could not find a version number
	if [[ -z "$VERSION" ]]; then
		echo "could not determine python version from: $PYVER"
		echo "please use a version like 3.10 or 3.12.1 or myenv3.11"
		exit 1
	fi

	# the environment already exists
	if [[ -d /ext3/miniconda3/envs/$CONDAENV ]]; then
		echo "using environment: $CONDAENV"
	else
		# create the environment

		echo "using environment: $CONDAENV"
		if [[ ! -d /ext3/miniconda3/envs/$CONDAENV ]]; then
			echo "creating environment: $CONDAENV"
			conda create -yq -n "$CONDAENV" python="$VERSION" pip

			# fail instructions
			if [[ ! $? -eq 0 ]]; then
				conda search python
				echo ""
				echo ""
				echo "failed to create environment: $CONDAENV with version $VERSION :("
				echo "please check the version against the available versions above"
				echo ""
				echo "Either rerun singuconda or you can just do it yourself! Do singrw and run:"
				echo ""
				echo "$ conda create -yq -n $CONDAENV python=$VERSION"
				echo "$ echo conda activate $CONDAENV > /ext3/conda.activate"
				echo ""
				exit 1
			fi
			echo "environment created: $CONDAENV"
		fi
	fi
fi

# add script to activate this environment

if [[ ! -z "$CONDAENV" ]]; then
	if [[ -d /ext3/miniconda3/envs/$CONDAENV ]]; then
		echo "conda activate $CONDAENV" > /ext3/conda.activate
	else 
		echo "environment $CONDAENV does not exist"
	fi
fi
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
