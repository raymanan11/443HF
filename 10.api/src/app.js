const express = require("express");
const app = express();
var morgan = require('morgan')
app.use(morgan('combined'))
const bodyparser = require("body-parser");
const { registerUser, userExist } = require("./registerUser");
const {createAsset,TransferAsset} =require('./tx')

const chaincodeName = "basic";
const channelName = "mychannel"

var cors = require('cors')
app.use(cors())
app.use(bodyparser.json());

app.listen(4000, () => {
    console.log("server started");

})

app.post("/register", async (req, res) => {

    try {
        let org = req.body.org;
        let userId = req.body.userId;
        let result = await registerUser({ OrgMSP: org, userId: userId });
        res.send(result);

    } catch (error) {
        res.status(500).send(error)
    }
});


app.post("/createAsset", async (req, res) => {
    try {


        let payload = {
            "org": req.body.org,
            "channelName": channelName,
            "chaincodeName": chaincodeName,
            "userId": req.body.userId,
            "data": req.body.data
        }

        let result = await createAsset(payload);
        res.send(result)
    } catch (error) {
        res.status(500).send(error)
    }
})

app.post("/transferAsset", async (req, res) => {

    try {

        let payload = {
            "org": req.body.org,
            "channelName": channelName,
            "chaincodeName": chaincodeName,
            "userId": req.body.userId,
            "data": req.body.data
        }

        let result = await TransferAsset(payload);
        res.send(result)
    } catch (error) {
        res.status(500).send(error)
    }
})



