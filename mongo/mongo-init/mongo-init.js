conn = new Mongo();
db = conn.getDB(process.env.MONGODB_DATABASE);
db.auth(process.env.MONGODB_USERNAME, process.env.MONGODB_PASSWORD)
db.log.insertOne({"message": "Database created."});
db.createCollection('event');



