export default class {

    static send401(res, message = 'Unauthorized') {
        res.status(401);
        return res.json({message: message});
    }

    static send403(res, message = 'Forbidden') {
        res.status(403);
        return res.json({message: message});
    }

    static send404(res, message = 'Ressource not found') {
        res.status(404);
        return res.json({message: message});
    }

    static send409(res, message = 'Conflict with ressource') {
        res.status(409);
        return res.json({message: message});
    }

    static send200(res, message = 'OK') {
        res.status(200);
        return res.json({message: message})
    }
}