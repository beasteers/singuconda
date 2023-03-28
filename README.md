# singuconda
Tool for setting up singularity overlays with miniconda

Here's the official NYU Greene documentation: https://sites.google.com/nyu.edu/nyu-hpc/hpc-systems/greene/software/singularity-with-miniconda

[singuconda.webm](https://user-images.githubusercontent.com/6741720/186782952-9a3b4a2c-5487-46a8-af21-786ba93fb6ed.webm)


## Install

```bash
ssh greene  # or whatever your environment is

curl -L https://github.com/beasteers/singuconda/raw/main/singuconda --output ~/singuconda
chmod +x ~/singuconda
```

## Tutorial


```bash
# cd to your projects directory
mkdir myproject
cd myproject

# make magic!
~/singuconda
```
The script will create some helper scripts for you:
 - `./sing` run the singularity container in read-only mode - use this to run many containers at once
 - `./singrw` run the singularity container in read-write mode - use this to install packages
 
By default it will auto-detect GPUs using nvidia-smi. But if that fails, you can do:
  - `./sing --nv`
  - `./singrw --nv`

Those commands above will create interactive sessions. If you want to run a script/commands in singularity, you can do this:

```bash
echo 'python  script.py' | ./sing

./sing <<< 'python  script.py'

./sing <<EOF
python script.py
EOF

./sing <<< "
python script.py
"
```

Any arguments you provide will be passed to the singularity command.

#### .gitignore

If you do this while you're inside a git repository, you may want to ignore the generated files. 

Here's a list of rules to filter them.
```
# the overlay file
*.ext3

# singuconda: start scripts
sing
singrw

# singuconda: named start scripts (for when you have multiple overlays in one directory)
sing-*
singrw-*

# the singularity container associated with the overlay
.*.sifpath
```

#### Uninstall

and if you want to remove the files, just do:

```bash
rm *.ext3 sing singrw sing-* singrw-* .*.sifpath
```

### Explanation

It will go through a series of prompts. What happens:
1. pick an overlay file
2. pick a sif file
3. install miniconda and allows you to select a specific python version if you want
4. adds the startup environment script (/ext3/env)
5. menu to install packages in the container
6. create shortcut script(s) for running the container

Then you're all done!

You can re-run it if you want to change anything (sif file, python version, installs).

This was built for NYU Greene's environment, but it should apply elsewhere too!

## Build

```bash
env GOOS=linux GOARCH=amd64 go build .
```

## Example SBatch file
So we have something to copy and paste from ;)
```bash
#!/bin/bash
#SBATCH -c 8
#SBATCH --mem 8GB
#SBATCH --time 8:00:00
#SBATCH --gres gpu:1
#SBATCH --job-name=myjob
#SBATCH --output logs/job.%J.out
#SBATCH --mail-type=ALL
#SBATCH --mail-user=<YOUR_USERID>@nyu.edu


../sing << EOF

python blah.py ...

EOF
```
