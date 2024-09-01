import nodeResolve from '@rollup/plugin-node-resolve'
import typescript from '@rollup/plugin-typescript'
import { defineConfig } from 'rollup'

export default defineConfig({
    input: 'src/main.ts',
    output: {
        dir: '../static/js',
        sourcemap: true,
        format: "iife",
    },
    plugins: [typescript(), nodeResolve()],
})
