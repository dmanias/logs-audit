conn = new Mongo();
db = conn.getDB(process.env.MONGODB_DATABASE);
db.auth(process.env.MONGODB_USERNAME, process.env.MONGODB_PASSWORD)
//db.log.insertOne({"message": "Database created."});
db.createCollection('events');
db.createCollection('users');
db.createCollection('authentication_tokens');

//db.users.createIndex({ "email": "dmanias@email.gr" }, { unique: true });
//db.users.insertOne({"username": "dmanias", "password": "1234" });