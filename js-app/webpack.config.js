const merge = require("webpack-merge");

const commonConfig = require('./config/webpack.common');
const developmentConfig = require('./config/webpack.dev');
const productionConfig = require('./config/webpack.prod');

module.exports = mode => {
    return merge(
        commonConfig,
        (mode === "production" ? productionConfig : developmentConfig),
        {mode}
    );
};