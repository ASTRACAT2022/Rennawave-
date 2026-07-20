import { Stack } from '@mantine/core'
import { SessionsExplorerWidget } from '@widgets/dashboard/sessions-explorer/sessions-explorer-widget'
import { DetailedUserInfoDrawerWidget } from '@widgets/dashboard/users/detailed-user-info-drawer/detailed-user-info-drawer.widget'
import { InternalSquadsDrawerWithStore } from '@widgets/dashboard/users/internal-squads-drawer-with-store'
import { UserAccessibleNodesModalWidget } from '@widgets/dashboard/users/user-accessible-nodes-modal/user-accessible-nodes.modal.widget'
import { UserTorrentBlockerReportsDrawerWidget } from '@widgets/dashboard/users/user-torrent-blocker-reports/user-torrent-blocker-reports.drawer.widget'
import { ViewUserModal } from '@widgets/dashboard/users/view-user-modal'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'

import { Page } from '@shared/ui'

export default function SessionsExplorerPageComponent() {
    const { t } = useTranslation()

    return (
        <Page title={t('constants.sessions-explorer')}>
            <Stack>
                <motion.div
                    animate={{ opacity: 1 }}
                    initial={{ opacity: 0 }}
                    transition={{ duration: 0.5, ease: [0, 0.71, 0.2, 1.01] }}
                >
                    <SessionsExplorerWidget />
                </motion.div>
            </Stack>

            <ViewUserModal key="view-user-modal" />
            <DetailedUserInfoDrawerWidget key="detailed-user-info-drawer" />
            <UserAccessibleNodesModalWidget key="user-accessible-nodes-modal" />
            <InternalSquadsDrawerWithStore key="internal-squads-drawer-with-store" />
            <UserTorrentBlockerReportsDrawerWidget key="user-torrent-blocker-reports-drawer" />
        </Page>
    )
}
