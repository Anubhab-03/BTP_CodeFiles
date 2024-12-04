package main

import (
    "context"
    "fmt"
    "sort"
    // "math/rand"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/rest"
    "net/http"
	// "os"
	"encoding/json"
	"strconv"
    "time"
    "io"
	// "errors"
//     "github.com/prometheus/client_golang/api"
//     v1 "github.com/prometheus/client_golang/api/prometheus/v1"
//     "github.com/prometheus/common/model"
// )
)

//promResponse struct for storing Prometheus' JSON response format
type promResponse struct {
    Status string `json:"status"`
    Data struct {
        ResultType string `json:"resultType"`
        Result []struct {
            Metric map[string]string `json:"metric"`
            Value  []interface{}     `json:"value"`
        } `json:"result"`
    } `json:"data"`
}


// Schedule pod to a node by setting its nodeName
func schedulePodToNode(clientset *kubernetes.Clientset, pod *corev1.Pod, nodeName string) error {
	binding := &corev1.Binding{
		ObjectMeta: metav1.ObjectMeta{
		    Name: pod.Name,
		    Namespace: pod.Namespace,
		},
		Target: corev1.ObjectReference{
		    Kind: "Node",
		    Name: nodeName,
		},
	    }
    err := clientset.CoreV1().Pods(pod.Namespace).Bind(context.TODO(), binding, metav1.CreateOptions{})
	if(err == nil){
		fmt.Printf("Scheduled pod %s to node %s\n",pod.Name,nodeName)
	}
    return err
}

func getNodeExporterInstance(nodeName string, clientset *kubernetes.Clientset) (string, error) {
    pods, err := clientset.CoreV1().Pods("monitoring").List(context.Background(), metav1.ListOptions{
        LabelSelector: "app.kubernetes.io/component=exporter,app.kubernetes.io/name=node-exporter",

    })
    if err != nil {
        return "", fmt.Errorf("Error fetching Node Exporter pods: %v", err)
    }

    for _, pod := range pods.Items {
        if pod.Spec.NodeName == nodeName {
            return pod.Status.PodIP + ":9100", nil // Pod IP with Node Exporter port
        }
    }

    return "", fmt.Errorf("No Node Exporter instance found for node: %s", nodeName)
}



// Function to get node metrics, a mock function for now
func getNodeMetrics(nodeName string,clientset *kubernetes.Clientset) float64 {
    instance, err := getNodeExporterInstance(nodeName, clientset)
    if err != nil {
        fmt.Printf("Error resolving Node Exporter instance: %v\n", err)
        return 0
    }

    // prometheusURL := "http://prometheus-service.monitoring.svc.cluster.local:8080/api/v1/query"
    prometheusURL := "http://localhost:8080/api/v1/query"
    // fmt.Println(nodeName)   

    // prometheusURL := "http://prometheus-service:8080/api/v1/query"
    query := fmt.Sprintf(`node_memory_MemFree_bytes{instance="%s"}`, instance)
	fmt.Println("Entered the get node metrics funciton")
    
    // Build the request
    req, err := http.NewRequest("GET", prometheusURL, nil)
    if err != nil {
        fmt.Printf("Error creating request: %v\n", err)
        return 0
    }

    // Add query parameters
    q := req.URL.Query()
    q.Add("query", query)
    req.URL.RawQuery = q.Encode()

    // Execute the request
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("Error querying Prometheus: %v\n", err)
        return 0
    }
    defer resp.Body.Close()

    // Read and parse the response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Error reading response: %v\n", err)
        return 0
    }

    var promResp promResponse
    if err := json.Unmarshal(body, &promResp); err != nil {
        fmt.Println("Entered condition 1")
        fmt.Printf("Error parsing JSON: %v\n", err)
        return 0
    }

    // fmt.Println("%v",len(promResp.Data.Result))

    // Extract and return the metric value
    if len(promResp.Data.Result) > 0 && len(promResp.Data.Result[0].Value) > 1 {
        value, ok := promResp.Data.Result[0].Value[1].(string)
        // fmt.Println("Value: %s",value)
        if ok {
            result, _ := strconv.ParseFloat(value, 64)
            fmt.Println("------------------------")
			fmt.Printf("Node name: %s\n", nodeName)
			fmt.Printf("Result: %f\n", result)			
            fmt.Println("------------------------")
            return result
            
        }
       
        fmt.Println("Not ok")
            
    }
    return 0
    
}

// Find the best node for a high-availability (HA) pod
func findBestNodeForPod(clientset *kubernetes.Clientset, pod *corev1.Pod) (string, error) {
	fmt.Println("Entered best node funciton")
    nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
     
    })
    if err != nil {
        return "", err
    }
    
    // Sort nodes by resource usage
    sort.Slice(nodes.Items, func(i, j int) bool {
        return getNodeMetrics(nodes.Items[i].Name,clientset) > getNodeMetrics(nodes.Items[j].Name,clientset)
    })
    
    if len(nodes.Items) > 0 {
        return nodes.Items[0].Name, nil
    }
    return "", fmt.Errorf("no suitable nodes found")
}

func getPodMemReq(pod *corev1.Pod) int64{
	if len(pod.Spec.Containers) > 0 {
        // Memory request of the first container
        if memReq, ok := pod.Spec.Containers[0].Resources.Requests[corev1.ResourceMemory]; ok {
			// fmt.Printf("Memory value of %s is %v\n",pod.Name,memReq.Value())
            return memReq.Value() // Returns memory in bytes
        }
    }
    return 0 // Default to 0 if no request is found
}

// Watch for unscheduled HA pods and schedule them
func watchForPodsAndSchedule(clientset *kubernetes.Clientset) {
    
    pods, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        panic(err.Error())
    }
    sort.Slice(pods.Items, func(i, j int) bool {
        return getPodMemReq(&pods.Items[i]) > getPodMemReq(&pods.Items[j])
    })
    startTime := time.Now()
    flag:= false
	for _, pod := range pods.Items {
        
        if pod.Spec.NodeName == "" { // Unscheduled pod
            flag = true
			nodeName, err := findBestNodeForPod(clientset, &pod)
			if err != nil {
				fmt.Printf("Failed to find node for pod %s: %v\n", pod.Name, err)
				continue
			}
			// Schedule the pod
			err = schedulePodToNode(clientset, &pod, nodeName)
			if err != nil {
				fmt.Printf("Failed to schedule pod %s: %v\n", pod.Name, err)
			}
            time.Sleep(15 * time.Second)
        }
    }
    
    endTime := time.Now()
    // fmt.Println("Start time:",startTime)
    // fmt.Println("End time:",endTime)
    elapsedTime := endTime.Sub(startTime)
    if flag == true{
        fmt.Println("Total time to scheduler pods: ",elapsedTime)
    }
}


func main() {
    
    config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
    if err != nil {
        config, err = rest.InClusterConfig()
        if err != nil {
            panic(err.Error())
        }
    }

    // Create a clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

    fmt.Println("Successfully connected to main_new Kubernetes API")

    // Start watching and scheduling HA pods
    for{
    watchForPodsAndSchedule(clientset)
    }
}

