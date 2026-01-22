#!/bin/bash
#SBATCH -c 8
#SBATCH --mem 24GB
#SBATCH --time 8:00:00
#SBATCH --gres gpu:1
#SBATCH --account=<YOUR PROJECT ID>
#SBATCH --job-name=jupyter
#SBATCH --output logs/jupyter.out

port=$(shuf -i 10000-65500 -n 1)

echo -e "\nTo access, run one of the following from your local machine:"
echo -e "ssh -L 8888:localhost:$port $USER@login.torch.hpc.nyu.edu"
echo -e "ssh -L 8888:localhost:$port torch"

echo -e "\nThen, from that shell, run:"
echo -e "ssh -L $port:localhost:$port $HOSTNAME"

echo -e "\nIn your local browser, open http://localhost:8888/"
echo -e "If asked for a token, copy the token from the URL in logs/jupyter.out\n"

./singrw << EOF

python -m ipykernel install --name sing --user
jupyter lab --no-browser --port $port

EOF