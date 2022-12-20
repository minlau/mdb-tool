const {merge} = require('webpack-merge');

module.exports = merge([{
    devtool: "inline-cheap-module-source-map",
}]);