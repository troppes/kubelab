import k8s from '@kubernetes/client-node';

const kc = new k8s.KubeConfig();
kc.loadFromFile(new URL('../kubeconfig', import.meta.url).pathname);

const k8sApi = kc.makeApiClient(k8s.CoreV1Api); // for now okay
const k8sApps = kc.makeApiClient(k8s.AppsV1Api);

export default class {

    static async scale(namespace, name, replicas) {
        // find the particular deployment
        const res = await k8sApps.readNamespacedDeployment(name, namespace);
        let deployment = res.body;
        // make changes
        deployment.spec.replicas = replicas;

        // updateResource
        
        return await (await k8sApps.replaceNamespacedDeployment(name, namespace, deployment)).response.statusCode;
    }
}

