const path = require('path');
const merge = require('webpack-merge');
const UglifyJsPlugin = require("terser-webpack-plugin");
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const OptimizeCSSAssetsPlugin = require("optimize-css-assets-webpack-plugin");

module.exports = merge([{
    entry: './src/index.js',
    module: {
        rules: [
            {
                test: /\.js$/,
                exclude: /node_modules/,
                use: "babel-loader?cacheDirectory"
            }, {
                test: /\.css$/,
                use: [
                    MiniCssExtractPlugin.loader,
                    // 'style-loader',
                    'css-loader'
                ]
            }, {
                test: /\.(png|jpg|gif|svg|eot|ttf|woff|woff2)$/,
                use: 'url-loader?limit=1024'
            }
        ]
    },
    optimization: {
        splitChunks: {
            cacheGroups: {
                commons: {
                    test: /[\\/]node_modules[\\/]/,
                    name: "vendor",
                    chunks: "initial",
                },
            },
        },
        minimizer: [
            new UglifyJsPlugin({
                cache: true,
                parallel: true,
                sourceMap: true // set to true if you want JS source maps
            }),
            new OptimizeCSSAssetsPlugin({})
        ]
    },
    output: {
        path: path.resolve(__dirname, "../../static/assets/"),
        publicPath: "assets/",
    },
    plugins: [
        new HtmlWebpackPlugin({template: './src/index.html', filename: '../index.html'}),
        new MiniCssExtractPlugin({
            filename: "main.min.css",
            chunkFilename: "main.chunk.css"
        })
    ]
}]);