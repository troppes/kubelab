import dao from './dao.js';

export default class {
    static async getAllUsers() {
        return await dao.all("SELECT * FROM users", []);
    }

    static async getUserById(id) {
        return dao.get('SELECT * FROM users WHERE id = ?', [id]);
    }

    static async getUserByName(name) {
        return dao.get('SELECT * FROM users WHERE name = ?', [name]);
    }

    static async addNewUser(name, hash, type) {
        return await dao.run("INSERT INTO users (name, password_hash, type) VALUES (?, ?, ?)", [name, hash, type]);
    }

    static async modifyUserById(id, newName = null, newPasswordHash = null, newType = null) {
        if (newName === null || newPasswordHash === null || newType == null) {
            try {
                const { name, password_hash, type } = await this.getUserById(id);
                newName = newName || name;
                newPasswordHash = newPasswordHash || password_hash;
                newType = newType || type;
            } catch (e) {
                console.log(e);
                return e;
            }
        }
        return await dao.run("UPDATE users SET name = ?, password_hash = ?, type = ? WHERE id = ?", [newName, newPasswordHash, newType, id]);
    }

    static async deleteUserById(id) {
        return await dao.run("DELETE FROM users WHERE id = ?", [id]);
    }

}