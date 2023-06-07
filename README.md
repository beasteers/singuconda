# singuconda 🌈
Tool for setting up singularity overlays with miniconda - [official NYU Greene docs](https://sites.google.com/nyu.edu/nyu-hpc/hpc-systems/greene/software/singularity-with-miniconda)

...because nobody likes doing it

---
✨ Here's what you could look like ✨

[singuconda.webm](https://user-images.githubusercontent.com/6741720/186782952-9a3b4a2c-5487-46a8-af21-786ba93fb6ed.webm)

---

Running singuconda will give you a magic little `./sing` 🧚🏾‍♀️ command in your current directory that:
 - autodetects GPUs and will automagically add the `--nv` flag
 - remembers the path to your overlay and sif images so all you have to do is `./sing`
 - It automatically sources your `env` file for you (the one from the tutorial)
 - also creates a `singrw` script that mounts the overlay in read-write mode (`./sing` mounts with read-only so you can have multiple scripts using it)
 - Has full support for both interactive shells (`./sing`) and scripts (`./sing <<< "type -P python"`) which will run the command and exit. This is what's used in sbatch files!
 - It accepts additional arguments so you can do `./sing -o /scratch/work/public/ml-datasets/coco/coco-2017.sqf:ro` to mount additional overlays (for example)


The `~/singuconda` script itself:
 - has autocomplete for all of the overlays and sif files
 - automatically installs miniconda and lets you optionally pick a python version



## Install

```bash
ssh greene  # or whatever your environment is

curl -L https://github.com/beasteers/singuconda/raw/main/singuconda --output ~/singuconda
chmod +x ~/singuconda
```

## Tutorial
The singuconda command should always be run from the directory where you want your overlay and sing script to live.

But once they're created, the `sing` script can be run from anywhere.

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
# These are no longer generated in the current version of the script
sing-*
singrw-*

# the singularity container associated with the overlay
.*.sifpath
```

### FAQ

##### I want to have two overlays in the same directory! How does `./sing` know which one to point to?

singuconda does allow creating multiple overlays in the same directory. 
When you use singuconda to setup a second overlay in the same directory, 
it will overwrite the sing command to point to your newer overlay. 

If you want to use your first overlay, you can override the overlay using `SING_NAME=my-first-sing ./sing`
(assuming your overlay is called `my-first-sing.ext3`). 

##### I renamed my overlay and now it's broken!!

Yep that's gonna happen if you do that! But don't fret.

You need to rename the hidden file that contains what SIF file you want to use.
```bash
OLD_SING_NAME=my-first-sing
NEW_SING_NAME=better-name

mv ".${OLD_SING_NAME}.sifpath" ".${NEW_SING_NAME}.sifpath"
```
And then you're going to want to edit `./sing` and `./singrw` to point to your new overlay name.

change
```bash
SING_NAME="${SING_NAME:-my-first-sing}"
```
to
```bash
SING_NAME="${SING_NAME:-better-name}"
```

##### I want to change my SIF file!

Just run `~/singuconda` again! It'll ask you if you want to configure an existing one or create a new one.

##### I need to choose a different overlay file (I didn't pick enough space)

Well that's a bummer! But I've done that too. Unfortunately, there's not a super convenient way, 
but fortunately it's very easy to just start over! (which is what I always do).

If you need to, I suppose you could try creating a new overlay, then mount both overlays and try to 
copy between them, but I'm not sure how to mount the second overlay
to a different directory (because afaik right now they'd both mount to `/ext3`). 

```
./singrw -o my-too-small-overlay.ext3  # uh oh! collision? I should test this lol
```

## Uninstall

and if you want to remove the files, just do:

```bash
rm *.ext3 sing singrw .*.sifpath
```

## Explanation

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

And for jupyter:

```bash
#!/bin/bash
#SBATCH -c 8
#SBATCH --mem 24GB
#SBATCH --time 8:00:00
#SBATCH --gres gpu:1
#SBATCH --job-name=jupyter
#SBATCH --output logs/jupyter.out

port=$(shuf -i 10000-65500 -n 1)
/usr/bin/ssh -N -f -R $port:localhost:$port log-1
/usr/bin/ssh -N -f -R $port:localhost:$port log-2
/usr/bin/ssh -N -f -R $port:localhost:$port log-3
echo "To access:"
echo "ssh -L $port:localhost:$port $USER@greene.hpc.nyu.edu"
echo "ssh -L $port:localhost:$port greene"

../singrw << EOF

python -m ipykernel install --name sing --user
jupyter lab --no-browser --port $port

EOF
```

## Extra Greene Helpers
Put your things in your home directory
```bash
ln -s /scratch/$USER ~/scratch
ln -s /vast/$USER ~/vast
ln -s /archive/$USER ~/archive
```

For your `~/.bashrc`:
```bash
# convenience commands for watching squeue
export SQUEUEFMT='%.18i %.9P %.32j %.8u %.8T %.10M %.9l %.6D %R'
alias msq='squeue  --me -o "$SQUEUEFMT"'
alias wsq='watch -n 2 "squeue --me -o \"$SQUEUEFMT\""'

# lets me know when my bashrc is sourced
[[ $- == *i* ]] && echo 'hi bea :)'
```
