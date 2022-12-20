const {merge} = require('webpack-merge');

module.exports = merge([{
    devtool: "nosources-source-map",

    performance: {
        hints: false
    }
}]);