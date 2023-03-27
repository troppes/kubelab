import repository from '../repositories/repository.js';
import statuscodes from "../lib/statuscodes.js";
import * as bcrypt from 'bcrypt';

const saltRounds = 10;

export default class {
    static async createUser(req, res) {
        try {
            const hashedPassword = await bcrypt.hash(req.body.password, saltRounds);
            await repository.addNewUser(req.body.name, hashedPassword, req.body.type);
            statuscodes.send200(res, 'User added successfully');
        } catch (e) {
            statuscodes.send409(res, e);
        }
    }

    static async deleteUser(req, res) {
        try {
            const machine = await repository.deleteUserById(req.params.id);
            if (machine['changes'] !== 0) {
                statuscodes.send200(res, 'User deleted successfully');
            } else {
                statuscodes.send404(res, 'User not found');
            }
        } catch (e) {
            statuscodes.send409(res, e);
        }
    }

    static async modifyUserById(req, res) {
        try {
            let hashedPassword = null;
            if(req.body.password != null) {
                hashedPassword = await bcrypt.hash(req.body.password, saltRounds);
            }
            const user = await repository.modifyUserById(req.params.id, req.body.name, hashedPassword, req.body.type);
            if (user.hasOwnProperty('changes')) {
                statuscodes.send200(res, 'User modified successfully');
            } else {
                statuscodes.send404(res, 'User not found');
            }
        } catch (e) {
            statuscodes.send409(res, e);
        }
    }

    static async listUsers(req, res) {
        try {
            let users = await repository.getAllUsers();
            return res.send({users});
        } catch (e) {
            statuscodes.send409(res, e);
        }
    }

    static async getUserById(req, res) {
        try {
            let user = await repository.getUserById(req.params.id);
            if (user === undefined) {
                statuscodes.send404(res, 'User not found');
            } else {
                return res.send({user});
            }
        } catch (e) {
            statuscodes.send409(res, e);
        }
    }
}
