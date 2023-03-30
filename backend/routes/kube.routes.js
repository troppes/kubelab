import * as express from 'express';
import kubeController from '../controllers/kube.controller.js';

const router = express.Router()

router.post("/scale/", kubeController.scale);

export default router;