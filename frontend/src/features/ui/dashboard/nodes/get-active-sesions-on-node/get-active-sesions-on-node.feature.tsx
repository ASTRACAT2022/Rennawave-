import { ActionIcon, Tooltip } from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import { NodeActiveSessionsDrawerWidget } from '@widgets/dashboard/nodes/node-active-sessions'
import { useTranslation } from 'react-i18next'
import { TbRadar } from 'react-icons/tb'

interface IProps {
    nodeUuid: string
}

export function GetActiveSessionsOnNodeFeature(props: IProps) {
    const { nodeUuid } = props
    const [activeSessionsModalOpened, activeSessionsModalHandlers] = useDisclosure(false)
    const { t } = useTranslation()

    return (
        <>
            <Tooltip label={t('get-user-usage.feature.active-sessions')}>
                <ActionIcon
                    color="indigo"
                    onClick={activeSessionsModalHandlers.open}
                    size="lg"
                    variant="soft"
                >
                    <TbRadar size="22px" />
                </ActionIcon>
            </Tooltip>

            <NodeActiveSessionsDrawerWidget
                nodeUuid={nodeUuid}
                onClose={activeSessionsModalHandlers.close}
                opened={activeSessionsModalOpened}
            />
        </>
    )
}
