from kubernetes import client, config
import random
# Configs can be set in Configuration class directly or using helper utility

def get_all_pods(api) -> list:
    response = api.list_pod_for_all_namespaces(watch=False)
    return response.items


def delete_random_pod(api, namespace) -> str:
    response = api.list_pod_for_all_namespaces(watch=False)
    pods = response.items
    pod_to_delete = random.choice(pods)
    api.delete_namespaced_pod(name=pod_to_delete.metadata.name, namespace=namespace)
    return pod_to_delete.metadata.name

def main():
    config.load_kube_config()
    api = client.CoreV1Api()

    pods = get_all_pods(api)
    # Listing All Pods
    print("IP\t\tNAMESPACE\tNAME")
    print("--------------------------------------------------")
    for i in pods:
        print("%s\t%s\t%s" % (i.status.pod_ip, i.metadata.namespace, i.metadata.name))

    # Deleting Random Pod
    namespace = "default"
    deleted_pod_name = delete_random_pod(api, namespace=namespace)
    print(f"Deleted Pod: {deleted_pod_name}")

if __name__ == "__main__":
    main()

