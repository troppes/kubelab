import Database from 'better-sqlite3';

const db = new Database('database/database.db', {verbose: console.log});

export default class {

    static all(stmt, params) {
        return new Promise((res, rej) => {
            try {
                return res(db.prepare(stmt).all(params));
            } catch (e) {
                return rej(e.message);
            }
        })
    }

    static get(stmt, params) {
        return new Promise((res, rej) => {
            try {
                return res(db.prepare(stmt).get(params));
            } catch (e) {
                return rej(e.message);
            }
        })
    }

    static run(stmt, params) {
        return new Promise((res, rej) => {
            try {
                return res(db.prepare(stmt).run(params));
            } catch (e) {
                return rej(e.message);
            }
        })
    }

}