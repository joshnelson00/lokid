from kubernetes import client, config
import random

def list_pods(pods: list) -> None:
    print("IP\t\tNAMESPACE\tNAME")
    print("--------------------------------------------------")
    for pod in pods:
        print("%s\t%s\t%s" % (pod.status.pod_ip, pod.metadata.namespace, pod.metadata.name))
    print("")
    return

def get_all_pods(api: client.CoreV1Api) -> list:
    response = api.list_pod_for_all_namespaces(watch=False)
    pods = response.items
    return pods

def get_pod_by_namespace(api: client.CoreV1Api, namespace: str) -> list:
    response = api.list_namespaced_pod(namespace=namespace)
    pods = response.items
    return pods

def delete_random_pod(api: client.CoreV1Api, namespace: str) -> str:
    response = api.list_pod_for_all_namespaces(watch=False)
    pods = response.items
    pod_to_delete = random.choice(pods)
    api.delete_namespaced_pod(name=pod_to_delete.metadata.name, namespace=namespace)
    print(f"Deleted Pod: {pod_to_delete.metadata.name}")
    return pod_to_delete

def list_nodes(nodes: list) -> None:
    print("NAME")
    print("-" * 20)
    for node in nodes:
        print("%s" % (node.metadata.name))
    print("")
    return

def get_all_nodes(api: client.CoreV1Api) -> list:
    response = api.list_node()
    nodes = response.items
    return nodes

def get_node_by_name(api: client.CoreV1Api, name: str) -> list:
    response = api.read_node(name=name)
    return response

def main():
    config.load_kube_config()
    api = client.CoreV1Api()

    # all_pods = get_all_pods(api)
    # list_pods(all_pods)
    
    nodes = get_all_nodes(api)
    list_nodes(nodes)
    
if __name__ == "__main__":
    main()