export const NODE_ADDRESS_PLACEHOLDER = '{{NODE_ADDRESS}}';

type AesingFlowInbound = {
    protocol?: unknown;
    streamSettings?: {
        tlsSettings?: {
            serverName?: unknown;
        };
    };
};

type XrayConfigWithInbounds = Record<string, unknown> & {
    inbounds?: AesingFlowInbound[];
};

/**
 * Produces a Node-specific config without mutating the shared profile config.
 * `{{NODE_ADDRESS}}` is intentionally supported only for AesingFlow TLS SNI.
 */
export function resolveNodeAddressPlaceholder(
    config: Record<string, unknown>,
    nodeAddress: string,
): Record<string, unknown> {
    const resolvedConfig = structuredClone(config) as XrayConfigWithInbounds;

    for (const inbound of resolvedConfig.inbounds ?? []) {
        if (inbound.protocol !== 'aesingflow') continue;

        const tlsSettings = inbound.streamSettings?.tlsSettings;
        if (tlsSettings?.serverName === NODE_ADDRESS_PLACEHOLDER) {
            tlsSettings.serverName = nodeAddress;
        }
    }

    return resolvedConfig;
}
