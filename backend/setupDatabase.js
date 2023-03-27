import Database from 'better-sqlite3';
import * as bcrypt from 'bcrypt';
import * as dotenv from 'dotenv';

dotenv.config()

const db = new Database('database/database.db', {verbose: console.log});
const saltRounds = 10;

const DEMO_USERS = [
    {username: 's1', password: 's1', type: 'student'},
    {username: 's2', password: 's2', type: 'student'}
]

db.exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT UNIQUE, password_hash TEXT, type TEXT)");


const userInsert = db.prepare('INSERT INTO users (name, password_hash, type) VALUES (?, ?, ?)');

const userTransactions = db.transaction((users) => {
    for (const user of users) {
        bcrypt.hash(user.password, saltRounds, function (err, hash) {
            userInsert.run(user.username, hash, user.type);
        });
    }
});

if (process.env.DEMO_DATA === 'TRUE') {
    userTransactions(DEMO_USERS);
}

bcrypt.hash(process.env.API_PASSWORD, saltRounds, function (err, hash) {
    userInsert.run(process.env.API_ADMIN, hash, 'admin');
});


