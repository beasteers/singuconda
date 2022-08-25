# singuconda
Tool for setting up singularity overlays with miniconda

Here's the official NYU Greene documentation: https://sites.google.com/nyu.edu/nyu-hpc/hpc-systems/greene/software/singularity-with-miniconda

## Install

```bash
ssh greene  # or whatever your environment is

curl https://github.com/beasteers/singuconda/raw/main/singuconda --output ~/singuconda
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
 
 And to run with GPU enabled, do:
  - `./sing --nv`
  - `./singrw --nv`

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
