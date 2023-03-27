import njwt from 'njwt';
import repository from '../repositories/repository.js';
import statuscodes from "../lib/statuscodes.js";

const {APP_SECRET = 'secret'} = process.env;

export const encodeToken = (tokenData) => {
    return njwt.create(tokenData, APP_SECRET).setExpiration().compact();
}

const decodeToken = (token) => {
    return njwt.verify(token, APP_SECRET);
}

export const authMiddleware = async (req, res, next) => {
    let token;
    if (req.header('Authorization')) {
        token = req.header('Authorization').split(" ")[1];
    }
    if (!token) {
        return next();
    }

    try {
        const decoded = decodeToken(token);
        const {userId} = decoded.body;
        const user = await repository.getUserById(userId)
        if (user) {
            req.userId = userId;
            req.userType = user.type;
        }
    } catch (e) {
        console.log("Token Error: " + e.message);
        return next();
    }

    next();
};

export const authenticated = (req, res, next) => {
    if (req.userId) {
        next();
    } else {
        statuscodes.send401(res, 'User not authenticated')
    }
}


