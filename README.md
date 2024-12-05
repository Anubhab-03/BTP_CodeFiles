## BTP setup

### Install docker 

1. sudo apt update
2. sudo apt install -y docker.io
3. sudo systemctl start docker
4. sudo systemctl enable docker

### Install kubectl

1. Download the latest release <br>
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

2. Install kubectl
   sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

3. Verify the version:- kubectl version --client

### Setting up minikube
1. Follow the steps given in the [official link](https://minikube.sigs.k8s.io/docs/start/?arch=%2Fwindows%2Fx86-64%2Fstable%2F.exe+download) to download minikube
   according to your system needs.

2. Start the cluster
   > minikube start

3. Tp specify the number of cpus and memory needed
   > minikube start --cpus = <no. of cpu> --memory=<mem. needed>

4. We can also view the minikube dashboard
   > minikube dashboard

