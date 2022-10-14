# install docker and docker-compose
sudo apt update
sudo apt install apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
apt-cache policy docker-ce
sudo apt install docker-ce
#docker compose
mkdir -p ~/.docker/cli-plugins/
curl -SL https://github.com/docker/compose/releases/download/v2.3.3/docker-compose-linux-x86_64 -o ~/.docker/cli-plugins/docker-compose
chmod +x ~/.docker/cli-plugins/docker-compose
#sudo usermod -aG docker ${USER}
#su - ${USER} -c "cd `pwd`; bash"
#install golang
sudo apt install golang-go
#Create Mongo directories
mkdir mongo/data
mkdir mongo/data/db
#chown ${USER}:${GROUP} mongo/data/db
#sudo chown -R $USER:$USER /mongo/data/db
#Run the service
#sudo docker compose build
#sudo docker compose up -d