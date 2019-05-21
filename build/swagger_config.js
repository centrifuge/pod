module.exports = {
    info: {
        version: "0.0.5",
        title: "Centrifuge OS Node API",
        description: "\n",
        contact: {
            name: "Centrifuge",
            url: "https://github.com/centrifuge/go-centrifuge",
            email: "hello@centrifuge.io",
        }
    },
    host: "localhost:8082",
    pathPrefix: "",
    schemes: ['http'],
    defaultSwagger: {
        consumes: ["application/json",],
        produces: ["application/json",],
    }
};
