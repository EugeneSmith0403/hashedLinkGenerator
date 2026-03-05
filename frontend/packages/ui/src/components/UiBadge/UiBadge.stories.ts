import type { Meta, StoryObj } from '@storybook/vue3'
import UiBadge from './UiBadge.vue'

const meta: Meta<typeof UiBadge> = {
  title: 'UI/Badge',
  component: UiBadge,
  tags: ['autodocs'],
  argTypes: {
    variant: { control: 'select', options: ['success', 'warning', 'danger', 'info', 'neutral'] },
  },
}

export default meta
type Story = StoryObj<typeof UiBadge>

export const Success: Story = {
  args: { variant: 'success' },
  render: (args) => ({ components: { UiBadge }, setup: () => ({ args }), template: '<UiBadge v-bind="args">Active</UiBadge>' }),
}

export const Danger: Story = {
  args: { variant: 'danger' },
  render: (args) => ({ components: { UiBadge }, setup: () => ({ args }), template: '<UiBadge v-bind="args">Canceled</UiBadge>' }),
}
