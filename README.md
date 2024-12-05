## Project setup

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

### Setting up prometheus
First create a separate namespace *monitoring* for deploying prometheus
> kubectl create namespace monitoring

  1. Create the prometheus configuration file - ***prometheus-config.yaml***
     > kubectl apply -f prometheus-config.yaml
  
  2. Create the prometheus service account file - ***prometheus-service.yaml***
     > kubectl apply -f prometheus-service.yaml
  
  3. Create the prometheus cluster role file - ***prometheus-cluster-role.yaml***
     > kubectl apply -f prometheus-cluster-role.yaml
  
  4. Create the prometheus deployment file - ***prometheus-deploy.yaml***
     > kubectl apply -f prometheus-deploy.yaml

Run the command - *kubectl get pods -o wide -n monitoring* to view the prometheus pod deployment

Now run the command :- 
> ***prometheus port-forward `<prometheus-pod-name>` 8080:9090 -n monitoring***

Now run *localhost:8080* in your web browser to view the prometheus server.

### Run the custom-scheduler I
1. Apply the high memory pod.
   > kubectl apply -f high_mem.yaml

2. Apply the low memory pod.
   > kubectl apply -f low_mem.yaml

3. Run the go file for custom scheduler I
   > go run scheduler_one.go

First, the high-mem pod gets scheduled and then the low_mem pod,irrespective of their order of deploying.
We can view in the terminal the scraped metric value of each node, and the best node out of them is selected for scheduling.

### Run the custom-scheduler II
1. Apply the high-availability pod.
   > kubectl apply -f high_mem.yaml

2. Apply the low-availability pod.
   > kubectl apply -f low_mem.yaml

3. Run the go file for custom scheduler II
   > go run scheduler_two.go

The high-avialbility pods are filtered first and are chosen for scheduling.
We can view in the terminal the scraped metric value of each node, and the best node out of them is selected for scheduling.

Run the command - *kubectl get pods -o wide* for viewing the deployed pods and their nodes.










