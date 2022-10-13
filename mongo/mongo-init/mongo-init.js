conn = new Mongo();
db = conn.getDB(process.env.MONGODB_DATABASE);
db.auth(process.env.MONGODB_USERNAME, process.env.MONGODB_PASSWORD)
db.createCollection('events');
db.createCollection('users');
db.createCollection('authentication_tokens');
db.events.createIndex( {service:1})
db.events.createIndex( {eventType:1})
db.events.createIndex( {timestamp:1})
db.events.createIndex( {tags:"text"})
db.events.createIndex( {data:"text"})
db.users.createIndex( {username:1, password: 1}, {unique: true})
db.users.createIndex( {username:1}, {unique: true})