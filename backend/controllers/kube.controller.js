import kubeRepo from '../repositories/kubernetes.js';
import statuscodes from "../lib/statuscodes.js";

export default class {
    static async scale(req, res) {
        try {
            const scaleStatusCode = await kubeRepo.scale(req.body.namespace, req.body.name, req.body.replicas);
            if (scaleStatusCode === 200) {
                statuscodes.send200(res, 'Ressource modified successfully');
            } else {
                statuscodes.send404(res, 'Ressource not found');
            }
        } catch (e) {
            console.error(e);
            statuscodes.send409(res, 'An error occured, please contact your administrator!');
        }
    }
}

