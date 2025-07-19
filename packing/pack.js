const fs = require('fs');
const path = require('path');
const htmlMinifier = require('html-minifier');
const { obfuscate } = require('javascript-obfuscator');
const archiver = require('archiver');
const CleanCSS = require('clean-css');


const outputDir = path.join(process.cwd(), 'build');
const sourceDir = path.resolve(process.cwd(), '..');
const allowedExtensions = ['.html', '.shtml', '.js', '.mjs', '.htaccess', '.css'];

const excludedJSFiles = ['tailwind.config.js'];

if (fs.existsSync(outputDir)) {
    fs.rmSync(outputDir, { recursive: true, force: true });
}
fs.mkdirSync(outputDir);

const obfuscationOptions = {
    target: 'browser',
    seed: Math.random() * 1000000,
    compact: true,
    controlFlowFlattening: true,
    controlFlowFlatteningThreshold: 1,
    deadCodeInjection: true,
    deadCodeInjectionThreshold: 1,
    debugProtection: true,
    disableConsoleOutput: true,
    identifierNamesGenerator: 'mangled-shuffled',
    identifiersPrefix: 'numelon',
    numbersToExpressions: true,
    renameGlobals: true,
    reservedNames: ["isValidEmail"],
    selfDefending: true,
    simplify: true,
    splitStrings: true,
    splitStringsChunkLength: 5,
    stringArray: true,
    stringArrayCallsTransform: true,
    stringArrayCallsTransformThreshold: 1,
    stringArrayEncoding: ['rc4'],
    stringArrayIndexShift: true,
    stringArrayRotate: true,
    stringArrayShuffle: true,
    stringArrayWrappersCount: 5,
    stringArrayWrappersType: 'function',
    stringArrayWrappersParametersMaxCount: 5,
    stringArrayWrappersChainedCalls: true,
    stringArrayIndexesType: ['hexadecimal-number'],
    transformObjectKeys: true
};

function shouldProcessFile(file) {
    const ext = path.extname(file).toLowerCase();
    return allowedExtensions.includes(ext) || allowedExtensions.includes(file); // Support for ".htaccess" (no extension)
}

function processFile(file) {
    const ext = path.extname(file).toLowerCase();
    const baseName = path.basename(file);
    const content = fs.readFileSync(file, 'utf8');
    const targetPath = path.join(outputDir, baseName);

    if ((ext === '.js' || ext === '.mjs') && !excludedJSFiles.includes(baseName)) {
        const obfuscated = obfuscate(content, obfuscationOptions).getObfuscatedCode();
        fs.writeFileSync(targetPath, obfuscated);
    } else if (ext === '.html' || ext === ".shtml") {
        const minified = htmlMinifier.minify(content, {
            removeComments: true,
            collapseWhitespace: true,
            minifyCSS: true,
            minifyJS: true,
            removeAttributeQuotes: true,
            removeEmptyAttributes: true
        });
        fs.writeFileSync(targetPath, minified);
    } else if (ext === '.css') {
        const minified = new CleanCSS({ level: 2 }).minify(content).styles;
        fs.writeFileSync(targetPath, minified);
    } else if (file.endsWith('.htaccess')) {
        fs.writeFileSync(targetPath, content);
    } else if (ext === '.js' || ext === '.mjs') {
        // copy unobfuscated js file directly
        fs.writeFileSync(targetPath, content);
    }
}

function createZipArchive() {
    return new Promise((resolve, reject) => {
        const output = fs.createWriteStream(path.join(outputDir, 'build.zip'));
        const archive = archiver('zip', { zlib: { level: 9 } });

        output.on('close', () => resolve());
        archive.on('error', err => reject(err));

        archive.pipe(output);
        fs.readdirSync(outputDir).forEach(file => {
            if (file !== 'build.zip') {
                const fullPath = path.join(outputDir, file);
                archive.file(fullPath, { name: file });
            }
        });
        // Add all files from build/
        archive.finalize();
    });
}

function cleanBuildExceptZip() {
    fs.readdirSync(outputDir).forEach(file => {
        if (file !== 'build.zip') {
            const fullPath = path.join(outputDir, file);
            fs.rmSync(fullPath, { recursive: true, force: true });
        }
    });
}

try {
    const files = fs.readdirSync(sourceDir, { withFileTypes: true });

    files.forEach(item => {
        const fullPath = path.join(sourceDir, item.name);
        if (item.isFile() && shouldProcessFile(item.name)) {
            processFile(fullPath);
        }
    });

    const shouldZip = process.argv.includes('--zip');

    if (shouldZip) {
        createZipArchive().then(() => {
            cleanBuildExceptZip();
            console.log('✔ Build completed and compressed to build.zip');
        }).catch(err => {
            console.error('❌ Failed to create ZIP:', err);
        });
    } else {
        console.log('✔ Build and obfuscation completed');
    }

} catch (err) {
    console.error('Build failed:', err);
}