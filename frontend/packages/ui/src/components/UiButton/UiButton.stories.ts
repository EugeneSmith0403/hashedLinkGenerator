import type { Meta, StoryObj } from '@storybook/vue3'
import UiButton from './UiButton.vue'

const meta: Meta<typeof UiButton> = {
  title: 'UI/Button',
  component: UiButton,
  tags: ['autodocs'],
  argTypes: {
    variant: { control: 'select', options: ['primary', 'secondary', 'danger', 'ghost'] },
    size: { control: 'select', options: ['sm', 'md', 'lg'] },
  },
}

export default meta
type Story = StoryObj<typeof UiButton>

export const Primary: Story = {
  args: { variant: 'primary' },
  render: (args) => ({ components: { UiButton }, setup: () => ({ args }), template: '<UiButton v-bind="args">Click me</UiButton>' }),
}

export const Secondary: Story = {
  args: { variant: 'secondary' },
  render: (args) => ({ components: { UiButton }, setup: () => ({ args }), template: '<UiButton v-bind="args">Cancel</UiButton>' }),
}

export const Danger: Story = {
  args: { variant: 'danger' },
  render: (args) => ({ components: { UiButton }, setup: () => ({ args }), template: '<UiButton v-bind="args">Delete</UiButton>' }),
}

export const Loading: Story = {
  args: { variant: 'primary', loading: true },
  render: (args) => ({ components: { UiButton }, setup: () => ({ args }), template: '<UiButton v-bind="args">Saving</UiButton>' }),
}
