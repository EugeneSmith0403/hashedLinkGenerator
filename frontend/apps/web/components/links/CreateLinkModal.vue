<script setup lang="ts">
import { useForm, useField } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { createLinkSchema } from '~/schemas/link'
import { useMutation, useQueryClient } from '@tanstack/vue-query'

defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const { t } = useI18n()
const queryClient = useQueryClient()

const { handleSubmit, errors, resetForm } = useForm({
  validationSchema: toTypedSchema(createLinkSchema),
})

const { value: url, handleBlur, handleChange } = useField<string>('url')

const { mutate, isPending, error } = useMutation({
  mutationFn: () => useLinksService().create(url.value),
  onSuccess() {
    queryClient.invalidateQueries({ queryKey: ['links'] })
    resetForm()
    emit('close')
  },
})

const onSubmit = handleSubmit(() => mutate())
</script>

<template>
  <UiModal :open="open" :title="t('dashboard.createLink')" @close="emit('close')">
    <form class="space-y-4" @submit.prevent="onSubmit">
      <UiInput
        id="url"
        v-model="url"
        :label="t('dashboard.url')"
        type="url"
        placeholder="https://example.com/very-long-url"
        required
        :error="errors.url"
        @blur="handleBlur"
        @change="handleChange"
      />
      <p v-if="error" class="text-sm text-red-600">{{ (error as Error).message }}</p>
    </form>
    <template #footer>
      <UiButton variant="secondary" @click="emit('close')">{{ t('common.cancel') }}</UiButton>
      <UiButton type="submit" :loading="isPending" @click="onSubmit">{{ t('common.create') }}</UiButton>
    </template>
  </UiModal>
</template>
