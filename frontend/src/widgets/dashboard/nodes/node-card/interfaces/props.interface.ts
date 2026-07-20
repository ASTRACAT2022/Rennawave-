import { GetAllNodesCommand } from '@remnawave/backend-contract'

export interface IProps {
    disableReordering?: boolean
    handleViewNode: (nodeUuid: string) => void
    isDragOverlay?: boolean
    isMobile: boolean
    node: GetAllNodesCommand.Response['response'][number]
}
