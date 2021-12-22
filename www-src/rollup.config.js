// main app and css processing
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import typescript from '@rollup/plugin-typescript';
import copy from 'rollup-plugin-copy';
import postCSS from 'rollup-plugin-postcss';
import livereload from 'rollup-plugin-livereload';
import {terser} from 'rollup-plugin-terser';
import preprocess from 'svelte-preprocess';
import svelte from 'rollup-plugin-svelte';

// special build configurations
import prodbuild from "./prodbuild";

const PROD = !process.env.ROLLUP_WATCH;
const OUT_DIR = "../www-build";
const SOURCE_STATIC_DIR = "public";
const SOURCE_DIR = "src";

console.log("Building:", (PROD ? "Production" : "Development"));

export default {
	input: `${SOURCE_DIR}/main.ts`,
	output: {
		sourcemap: !PROD,
		format: 'iife',
		name: 'app',
		file: `${OUT_DIR}/static/app.[hash].js`
	},
	plugins: [
		copy({
			targets: [
				{ src: [`${SOURCE_STATIC_DIR}/static`,
					], dest: `${OUT_DIR}/` },
				{ src: [`${SOURCE_STATIC_DIR}/index.html`,
						`${SOURCE_STATIC_DIR}/favicon.ico`,
					], dest: `${OUT_DIR}/` }
			],
			verbose: true
		}),

		svelte({
			compilerOptions: {
				// enable run-time checks when not in production
				dev: !PROD
			},

			extensions: [".svelte"],

			preprocess: [
				preprocess()
			],
		}),

		// we'll extract any component CSS out into
		// a separate file - better for performance
		postCSS({
			extract: true,
			sourceMap: !PROD,
		}),

		// If you have external dependencies installed from
		// npm, you'll most likely need these plugins. In
		// some cases you'll need additional configuration -
		// consult the documentation for details:
		// https://github.com/rollup/plugins/tree/master/packages/commonjs
		resolve({
			browser: true,
			dedupe: ['svelte']
		}),

		commonjs(),

		typescript({
			sourceMap: !PROD,
			inlineSources: !PROD
		}),
		
		// In dev mode, call `npm run start` once
		// the bundle has been generated
		!PROD && serve(),

		// Watch the `public` directory and refresh the
		// browser on changes when not in production
		!PROD && livereload({
			watch: [`${OUT_DIR}/`],
			verbose: true,
		}),

		// If we're building for production (npm run build
		// instead of npm run dev), minify
		PROD && terser(),
		PROD && prodbuild(OUT_DIR),
	],
	watch: {
		clearScreen: false
	}
};

function serve() {
	let server;

	function toExit() {
		if (server) server.kill(0);
	}

	return {
		writeBundle() {
			if (server) return;
			server = require('child_process').spawn('npm', ['run', 'start', '--', '--dev'], {
				stdio: ['ignore', 'inherit', 'inherit'],
				shell: true
			});

			process.on('SIGTERM', toExit);
			process.on('exit', toExit);
		}
	};
}