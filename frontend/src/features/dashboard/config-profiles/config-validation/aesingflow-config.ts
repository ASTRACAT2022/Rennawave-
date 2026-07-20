type JsonObject = Record<string, unknown>

const REMNAWAVE_SSL_DIRECTORY = '/var/lib/remnawave/configs/xray/ssl/'

const isObject = (value: unknown): value is JsonObject =>
    typeof value === 'object' && value !== null && !Array.isArray(value)

const pushConst = (schema: unknown, value: string): void => {
    if (!isObject(schema) || !Array.isArray(schema.anyOf)) return
    if (!schema.anyOf.some((item) => isObject(item) && item.const === value)) {
        schema.anyOf.push({ const: value })
    }
}

/**
 * Extends the upstream Xray schema in memory. The deployed validator assets
 * remain untouched; this only documents the custom AesingFlow extension that
 * is supported by Remnawave Nodes running the matching custom Xray core.
 */
export const extendAesingFlowSchema = <TSchema extends JsonObject>(schema: TSchema): TSchema => {
    const definitions = schema.definitions
    if (!isObject(definitions)) return schema

    const inbound = definitions.InboundObject
    const streamSettings = definitions.StreamSettingsObject
    const inboundSettings = definitions.InboundConfigurationObject

    if (!isObject(inbound) || !isObject(streamSettings) || !isObject(inboundSettings)) {
        return schema
    }

    const inboundProperties = inbound.properties
    const streamProperties = streamSettings.properties
    if (!isObject(inboundProperties) || !isObject(streamProperties)) return schema

    pushConst(inboundProperties.protocol, 'aesingflow')
    pushConst(streamProperties.network, 'aesingflow')

    const aesingFlowSettingsDefinition = 'AesingFlowInboundSettingsObject'
    definitions[aesingFlowSettingsDefinition] = {
        title: 'AesingFlow inbound settings',
        type: 'object',
        properties: {
            clients: {
                type: 'array',
                items: { type: 'object' }
            },
            maxStreams: { type: 'integer', minimum: 1 },
            congestionControl: { enum: ['brutal', 'cubic'] },
            brutalBps: { type: 'integer', minimum: 1 }
        },
        required: ['clients'],
        additionalProperties: false
    }

    const anyOf = inboundSettings.anyOf
    if (
        Array.isArray(anyOf) &&
        !anyOf.some(
            (item) =>
                isObject(item) && item.$ref === `#/definitions/${aesingFlowSettingsDefinition}`
        )
    ) {
        anyOf.push({ $ref: `#/definitions/${aesingFlowSettingsDefinition}` })
    }

    const allOf = Array.isArray(inbound.allOf) ? inbound.allOf : []
    if (!allOf.some((item) => isObject(item) && item.$comment === 'remnawave-aesingflow')) {
        allOf.push({
            $comment: 'remnawave-aesingflow',
            if: {
                properties: { protocol: { const: 'aesingflow' } },
                required: ['protocol']
            },
            then: {
                properties: {
                    settings: { $ref: `#/definitions/${aesingFlowSettingsDefinition}` },
                    streamSettings: {
                        allOf: [
                            { $ref: '#/definitions/StreamSettingsObject' },
                            {
                                type: 'object',
                                properties: {
                                    network: { const: 'aesingflow' },
                                    security: { const: 'tls' },
                                    tlsSettings: {
                                        type: 'object',
                                        properties: {
                                            serverName: { type: 'string', minLength: 1 },
                                            minVersion: { const: '1.3' },
                                            certificates: {
                                                type: 'array',
                                                minItems: 1,
                                                items: {
                                                    type: 'object',
                                                    properties: {
                                                        keyFile: {
                                                            type: 'string',
                                                            pattern:
                                                                '^/var/lib/remnawave/configs/xray/ssl/.+\\.key$'
                                                        },
                                                        certificateFile: {
                                                            type: 'string',
                                                            pattern:
                                                                '^/var/lib/remnawave/configs/xray/ssl/.+\\.pem$'
                                                        }
                                                    },
                                                    required: ['keyFile', 'certificateFile']
                                                }
                                            }
                                        },
                                        required: ['serverName', 'minVersion', 'certificates']
                                    }
                                },
                                required: ['network', 'security', 'tlsSettings']
                            }
                        ]
                    }
                },
                required: ['settings', 'streamSettings']
            }
        })
        inbound.allOf = allOf
    }

    return schema
}

/** Returns a semantic error for the custom inbound, or null when it is valid. */
export const validateAesingFlowConfig = (config: unknown): string | null => {
    if (!isObject(config) || !Array.isArray(config.inbounds)) return null

    for (const inbound of config.inbounds) {
        if (!isObject(inbound) || inbound.protocol !== 'aesingflow') continue

        const settings = inbound.settings
        const streamSettings = inbound.streamSettings
        if (!isObject(settings) || !Array.isArray(settings.clients)) {
            return 'AesingFlow requires settings.clients.'
        }
        if (!isObject(streamSettings) || streamSettings.network !== 'aesingflow') {
            return 'AesingFlow requires streamSettings.network = "aesingflow".'
        }
        if (streamSettings.security !== 'tls') {
            return 'AesingFlow requires streamSettings.security = "tls".'
        }
        const tlsSettings = streamSettings.tlsSettings
        if (!isObject(tlsSettings) || tlsSettings.minVersion !== '1.3') {
            return 'AesingFlow requires tlsSettings.minVersion = "1.3".'
        }
        if (typeof tlsSettings.serverName !== 'string' || tlsSettings.serverName.length === 0) {
            return 'AesingFlow requires tlsSettings.serverName.'
        }
        if (!Array.isArray(tlsSettings.certificates) || tlsSettings.certificates.length === 0) {
            return 'AesingFlow requires tlsSettings.certificates.'
        }
        for (const certificate of tlsSettings.certificates) {
            if (!isObject(certificate)) return 'AesingFlow certificate entry is invalid.'
            const keyFile = certificate.keyFile
            const certificateFile = certificate.certificateFile
            if (
                typeof keyFile !== 'string' ||
                !keyFile.startsWith(REMNAWAVE_SSL_DIRECTORY) ||
                !keyFile.endsWith('.key') ||
                keyFile.includes('..') ||
                typeof certificateFile !== 'string' ||
                !certificateFile.startsWith(REMNAWAVE_SSL_DIRECTORY) ||
                !certificateFile.endsWith('.pem') ||
                certificateFile.includes('..')
            ) {
                return 'AesingFlow certificates must use .key/.pem files inside the Remnawave SSL directory.'
            }
        }
    }
    return null
}

/** Removes custom inbounds before passing the remaining standard config to the existing Xray validator. */
export const withoutAesingFlowInbounds = (config: JsonObject): JsonObject => ({
    ...config,
    inbounds: Array.isArray(config.inbounds)
        ? config.inbounds.filter(
              (inbound) => !isObject(inbound) || inbound.protocol !== 'aesingflow'
          )
        : config.inbounds
})
