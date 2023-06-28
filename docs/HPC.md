# NYU Greene Crash Course

So you have an HPC account! 

#### Where do I start?

These are the main places to work out of.

```bash
# temporary storage where you should do most of your work
# put code and data here (but remember to have backups because these files can be purged)
echo $SCRATCH  # /scratch/$USER

# fast-access, can handle large data/inodes but space is limited so don't use this for long-term storage
echo $VAST  # /vast/$USER

# long-term storage - you should zip/tar/squashfs files before putting them here
echo $ARCHIVE  # /archive/$USER
```

I like to symlink them to my home directory so I can do `cd ~/scratch`

```bash
ln -s /scratch/$USER ~/scratch
ln -s /vast/$USER ~/vast
ln -s /archive/$USER ~/archive
```

#### How do I run my code interactively?
Well first thing to know is that when you first log into HPC, you're on what's called a "log" or "login" node. 
They aren't meant to run anything computationally heavy. So if you start typing `python ...` and your shell says something like `log-2`,
you should submit a job request to the cluster instead.

When I'm running something small, I'll run something like this to get an interactive shell:
```bash
# 1 cpu, for 3 hours, with 8GB of memory. 
srun -c 1 -t 3:0:0 --mem 8GB --pty bash -i
```

When I'm debugging a model and need a GPU, I might scale up to something like:
```bash
# 8 cpus, for 3 hours, with 32GB of memory, and 1 GPU. 
srun -c 8 -t 3:0:0 --mem 32GB --gres gpu:1 --pty bash -i
```
I believe the max CPUs you can use is 48. The max GPUs is 4 (though I think there's some nuance I'm missing here?). Max time is 1 week (`7-0:0:0`).

But try not to over-request resources you don't need because it takes resources away from others and it makes your jobs take longer to be allocated. 
Also, a common issue is jobs being killed for underutilizing GPUs so try not to request GPUs you're not using.

To see the jobs I have running, run this:
```bash
squeue --me
```

I also added these lines to my `~/.bashrc` file which gives me some commands with better outputs.
```bash
export SQUEUEFMT='%.18i %.9P %.32j %.8u %.8T %.10M %.9l %.6D %R'

# print out my jobs once
alias msq='squeue  --me -o "$SQUEUEFMT"'

# watch my jobs continuously, refresh every 2 seconds
alias wsq='watch -n 2 "squeue --me -o \"$SQUEUEFMT\""'
```
So now you can run `msq` or `wsq` instead.

If you want to open another terminal into the node, you can also SSH into the nodes that your jobs are submitted on by looking at the NODELIST column. 

```bash
[bs3639@log-1 ~]$ srun -c 1 -t 1:0:0 --mem 16GB --pty bash -i
srun: job 34999294 queued and waiting for resources
srun: job 34999294 has been allocated resources
hi bea :)
bash-4.4$ msq
             JOBID PARTITION                             NAME     USER    STATE       TIME TIME_LIMI  NODES NODELIST(REASON)
          34999294        cm                             bash   bs3639  RUNNING       0:11   1:00:00      1 cm002
bash-4.4$ ssh cm002
hi bea :)
[bs3639@cm002 ~]$ 
```

#### How do I run python?
Assuming you've installed singuconda:

```bash
cd ~/scratch
mkdir my-project
cd my-project

# get your code
git clone ...

# create your overlay
~/singuconda

# go into singularity (in write mode) to install any packages
./singrw
> pip install ...
> exit


```

#### How do I submit batch jobs?


