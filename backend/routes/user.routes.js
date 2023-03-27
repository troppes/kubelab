import userController from '../controllers/user.controller.js';
import * as express from 'express';

const router = express.Router()

router.get("/", userController.listUsers);
router.put("/", userController.createUser);
router.get("/id/:id", userController.getUserById);
router.post("/id/:id", userController.modifyUserById);
router.delete("/id/:id",userController.deleteUser);

export default router;