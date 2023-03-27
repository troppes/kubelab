import repository from '../repositories/repository.js';
import {encodeToken} from '../middleware/auth.middleware.js';
import statuscodes from "../lib/statuscodes.js";
import * as bcrypt from 'bcrypt';

const basicAuthDecrypt = (authString) => {
    let userData = Buffer.from(authString.split(" ")[1], 'base64').toString().split(':');

    return {name: userData[0], password_hash: userData[1]}
}

export const login = async (req, res) => {
    let {name, password_hash} = basicAuthDecrypt(req.headers.authorization);
    const user = await repository.getUserByName(name)

    if (!user) {
        statuscodes.send401(res, 'Invalid username or password');
    } else {
        bcrypt.compare(password_hash, user.password_hash, (err, result) => {
            if (result) {
                const accessToken = encodeToken({userId: user.id, type: user.type});
                return res.json({accessToken});
            } else {
                return statuscodes.send401(res, 'Invalid username or password');
            }
        });
    }
}