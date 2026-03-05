import {
  UiButton,
  UiInput,
  UiBadge,
  UiCard,
  UiModal,
  UiTable,
  UiSpinner,
  UiTooltip,
} from '@repo/ui'

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp
    .component('UiButton', UiButton)
    .component('UiInput', UiInput)
    .component('UiBadge', UiBadge)
    .component('UiCard', UiCard)
    .component('UiModal', UiModal)
    .component('UiTable', UiTable)
    .component('UiSpinner', UiSpinner)
    .component('UiTooltip', UiTooltip)
})
