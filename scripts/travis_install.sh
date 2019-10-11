#!/bin/bash

set -e -u -x

wget --quiet https://repo.anaconda.com/miniconda/Miniconda3-4.5.11-Linux-x86_64.sh -O ~/miniconda.sh

/bin/bash ~/miniconda.sh -b -p ~/miniconda
~/miniconda/bin/conda clean -tipsy
echo ". ~/miniconda/etc/profile.d/conda.sh" >> ~/.bashrc
echo "conda activate base" >> ~/.bashrc

. ~/miniconda/etc/profile.d/conda.sh
conda activate base
conda config --set auto_update_conda False
conda config --add channels https://dnap.cog.sanger.ac.uk/npg/conda/devel/generic
conda create -y -n travis
conda activate travis
conda install -y irods-icommands
conda install -y baton=2.0.0

mkdir -p ~/.irods
cat <<EOF > ~/.irods/irods_environment.json
{
    "irods_host": "localhost",
    "irods_port": 1247,
    "irods_user_name": "irods",
    "irods_zone_name": "testZone",
    "irods_home": "/testZone/home/irods",
    "irods_plugins_home": "$HOME/miniconda/envs/travis/lib/irods/plugins/",
    "irods_default_resource": "testResc"
}
EOF
