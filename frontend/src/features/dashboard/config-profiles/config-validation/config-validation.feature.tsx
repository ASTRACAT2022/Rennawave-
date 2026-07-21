import type { editor } from 'monaco-editor'

import { GetSnippetsCommand } from '@remnawave/backend-contract'
import consola from 'consola/browser'
import dayjs from 'dayjs'
import { RefObject } from 'react'

import { validateAesingFlowConfig, withoutAesingFlowInbounds } from './aesingflow-config'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const replaceSnippetsInArray = (array: any[], snippetsMap: Map<string, unknown>): void => {
    for (let i = array.length - 1; i >= 0; i--) {
        const item = array[i]

        if (item.snippet) {
            const snippet = snippetsMap.get(item.snippet)

            if (snippet) {
                if (Array.isArray(snippet)) {
                    array.splice(i, 1, ...snippet)
                } else {
                    // eslint-disable-next-line no-param-reassign
                    array[i] = snippet
                }
            } else {
                consola.error(`Snippet ${item.snippet} not found`)
                array.splice(i, 1)
            }
        }
    }
}

export const ConfigValidationFeature = {
    validate: (
        editorRef: RefObject<editor.IStandaloneCodeEditor | null>,

        setResult: (message: string) => void,
        setIsConfigValid: (isValid: boolean) => void,
        snippetsMap: Map<
            string,
            GetSnippetsCommand.Response['response']['snippets'][number]['snippet']
        >
    ) => {
        try {
            if (!editorRef.current) return

            const currentValue = editorRef.current.getValue()

            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            let clonedCurrentValue: any
            try {
                clonedCurrentValue = JSON.parse(currentValue)
            } catch {
                setResult(`${dayjs().format('HH:mm:ss')} | Invalid JSON.`)
                setIsConfigValid(false)
                return
            }

            const aesingFlowError = validateAesingFlowConfig(clonedCurrentValue)
            if (aesingFlowError) {
                setResult(`${dayjs().format('HH:mm:ss')} | ${aesingFlowError}`)
                setIsConfigValid(false)
                return
            }

            if (clonedCurrentValue.outbounds) {
                replaceSnippetsInArray(clonedCurrentValue.outbounds, snippetsMap)
            }

            if (clonedCurrentValue.routing?.rules) {
                replaceSnippetsInArray(clonedCurrentValue.routing.rules, snippetsMap)
            }

            if (clonedCurrentValue.routing?.balancers) {
                replaceSnippetsInArray(clonedCurrentValue.routing.balancers, snippetsMap)
            }

            const hasAesingFlowInbound =
                Array.isArray(clonedCurrentValue.inbounds) &&
                clonedCurrentValue.inbounds.some(
                    (inbound: unknown) =>
                        typeof inbound === 'object' &&
                        inbound !== null &&
                        (inbound as { protocol?: unknown }).protocol === 'aesingflow'
                )

            const xrayParseConfig = window.XrayParseConfig
            if (typeof xrayParseConfig !== 'function') {
                if (hasAesingFlowInbound) {
                    setResult(
                        `${dayjs().format('HH:mm:ss')} | AesingFlow config is valid. ` +
                            'Standard Xray WASM validation is temporarily unavailable.'
                    )
                    setIsConfigValid(true)
                    return
                }

                setResult(`${dayjs().format('HH:mm:ss')} | Xray WASM validator is unavailable. Restart it and try again.`)
                setIsConfigValid(false)
                return
            }

            const validationInput = withoutAesingFlowInbounds(clonedCurrentValue)
            const validationResult = xrayParseConfig(JSON.stringify(validationInput))

            const successMessage = hasAesingFlowInbound
                ? 'Xray config is valid. AesingFlow requires an AesingFlow-capable custom core on every Node.'
                : 'Xray config is valid.'

            setResult(`${dayjs().format('HH:mm:ss')} | ${validationResult || successMessage}`)
            setIsConfigValid(!validationResult)
        } catch (err: unknown) {
            const message = (err as Error).message
            if (message?.includes('Go program has already exited')) {
                setResult(`${dayjs().format('HH:mm:ss')} | WASM module crashed, restarting...`)
            } else {
                setResult(`${dayjs().format('HH:mm:ss')} | Validation error: ${message}`)
            }
            setIsConfigValid(false)
        }
    }
}
