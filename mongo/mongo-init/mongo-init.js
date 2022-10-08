conn = new Mongo();
db = conn.getDB(process.env.MONGODB_DATABASE);
db.auth(process.env.MONGODB_USERNAME, process.env.MONGODB_PASSWORD)
db.log.insertOne({"message": "Database created."});


//id: ID!
///name: String!
//banner:
//  base64: String
//  url: String
//homepage: String
//location:
//  street: String!
//  number: String!
//  town: String!
//  coordinates:
///    latitude: String
//    longitude: String
//telephone: [Int]

db.createCollection('shop');
db.shop.createIndex({ "afm": 1 }, { unique: true });
db.shop.insertOne({ "address": { "city": "Paris", "zip": "123" }, "afm":345, "name": "Shop1", "phone": "1234" });
db.shop.insertOne({ "address": { "city": "Marsel", "zip": "321" }, "afm":435, "name": "Shop2", "phone": "4321" });

db.createCollection('item');
db.item.insertOne({  "cost":12.5, "name": "Peponi", "category": "fruit" });
db.item.insertOne({  "cost":12, "name": "Karpouzi", "category": "fruit" });
db.item.insertOne({  "cost":19, "name": "Mousakas", "category": "food" });

db.createCollection('user');
db.user.insertOne({ "name": "Pantelis", "category": "admin" });
db.user.insertOne({ "name": "Dimos", "category": "customer" });




