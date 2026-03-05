import type { Meta, StoryObj } from '@storybook/vue3'
import UiInput from './UiInput.vue'

const meta: Meta<typeof UiInput> = {
  title: 'UI/Input',
  component: UiInput,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof UiInput>

export const Default: Story = {
  args: { label: 'Email', placeholder: 'you@example.com', type: 'email' },
  render: (args) => ({ components: { UiInput }, setup: () => ({ args }), template: '<UiInput v-bind="args" />' }),
}

export const WithError: Story = {
  args: { label: 'Email', modelValue: 'bad', error: 'Invalid email address' },
  render: (args) => ({ components: { UiInput }, setup: () => ({ args }), template: '<UiInput v-bind="args" />' }),
}

export const Disabled: Story = {
  args: { label: 'Read-only', modelValue: 'locked@example.com', disabled: true },
  render: (args) => ({ components: { UiInput }, setup: () => ({ args }), template: '<UiInput v-bind="args" />' }),
}
