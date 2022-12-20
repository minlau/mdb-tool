const {merge} = require('webpack-merge');

const commonConfig = require('./config/webpack.common');
const developmentConfig = require('./config/webpack.dev');
const productionConfig = require('./config/webpack.prod');

module.exports = webpackData => {
    let mode = webpackData.production === true ? "production" : "development"
    return merge(
        commonConfig,
        (mode === "production" ? productionConfig : developmentConfig),
        {mode}
    );
};