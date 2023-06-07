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

./singrw << EOF

python -m ipykernel install --name sing --user
jupyter lab --no-browser --port $port

EOF