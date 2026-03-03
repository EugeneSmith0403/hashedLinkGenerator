<template>
  <div class="center">
    <div class="flow-card">

      <!-- Stepper -->
      <div class="stepper">
        <div
          v-for="(s, i) in steps"
          :key="i"
          class="step"
          :class="{
            active: currentStep === i,
            done: currentStep > i,
          }"
        >
          <div class="step-circle">
            <span v-if="currentStep > i">✓</span>
            <span v-else>{{ i + 1 }}</span>
          </div>
          <span class="step-label">{{ s }}</span>
        </div>
      </div>

      <!-- Step 0: Stripe key config -->
      <section v-if="currentStep === 0">
        <h2>Stripe Setup</h2>
        <p class="desc">Enter your Stripe Publishable Key to initialize the card form.</p>

        <div class="field">
          <label>Publishable Key</label>
          <input
            v-model="stripeKey"
            type="text"
            placeholder="pk_test_..."
          />
        </div>

        <div v-if="keyError" class="alert error">{{ keyError }}</div>

        <button class="btn-primary" @click="initStripe">Continue</button>
      </section>

      <!-- Step 1: Add payment method -->
      <section v-if="currentStep === 1">
        <h2>Add Payment Method</h2>
        <p class="desc">Enter your card details. They are sent directly to Stripe — your server never sees raw card data.</p>

        <div class="field">
          <label>Card Details</label>
          <div class="stripe-wrapper" :class="{ focused: cardFocused }" id="card-element"></div>
        </div>

        <div v-if="addCardMutation.isError.value" class="alert error">
          {{ addCardMutation.error.value?.message }}
        </div>
        <div v-if="stripeError" class="alert error">{{ stripeError }}</div>

        <button
          class="btn-primary"
          :disabled="addCardMutation.isPending.value || confirmingCard"
          @click="handleAddCard"
        >
          {{ addCardMutation.isPending.value || confirmingCard ? 'Processing...' : 'Save Card' }}
        </button>
      </section>

      <!-- Step 2: Choose plan & subscribe -->
      <section v-if="currentStep === 2">
        <h2>Choose Plan</h2>
        <p class="desc">Select a plan and confirm your subscription.</p>

        <div class="plans">
          <label
            v-for="plan in plans"
            :key="plan.id"
            class="plan-card"
            :class="{ selected: selectedPlanId === plan.id }"
          >
            <input type="radio" :value="plan.id" v-model="selectedPlanId" hidden />
            <div class="plan-name">{{ plan.name }}</div>
            <div class="plan-price">${{ plan.cost }}<span>/mo</span></div>
            <ul class="plan-features">
              <li v-for="f in plan.features" :key="f">{{ f }}</li>
            </ul>
          </label>
        </div>

        <div class="field" style="margin-top: 20px;">
          <label>Plan ID (manual override)</label>
          <input v-model.number="selectedPlanId" type="number" min="1" placeholder="1" />
          <div class="hint">Enter the plan ID from your database</div>
        </div>

        <div v-if="subscribeMutation.isError.value" class="alert error">
          {{ subscribeMutation.error.value?.message }}
        </div>

        <button
          class="btn-primary"
          :disabled="!selectedPlanId || subscribeMutation.isPending.value"
          @click="handleSubscribe"
        >
          {{ subscribeMutation.isPending.value ? 'Creating subscription...' : `Subscribe to Plan #${selectedPlanId || '?'}` }}
        </button>
      </section>

      <!-- Step 3: Success -->
      <section v-if="currentStep === 3" class="success-section">
        <div class="success-icon">✓</div>
        <h2>Subscription Active!</h2>
        <p class="desc">Your subscription has been created successfully.</p>

        <div v-if="subscription" class="sub-details">
          <div class="detail-row">
            <span>Subscription ID</span>
            <span>{{ subscription.id }}</span>
          </div>
          <div class="detail-row">
            <span>Status</span>
            <span class="status-badge" :class="subscription.status">{{ subscription.status }}</span>
          </div>
          <div class="detail-row">
            <span>Customer</span>
            <span>{{ subscription.customer }}</span>
          </div>
        </div>

        <button class="btn-secondary" @click="resetFlow">Start Over</button>
      </section>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useMutation } from '@tanstack/vue-query'
import { loadStripe } from '@stripe/stripe-js'
import { subscriptionApi } from '../api/subscription.js'

// ── Hardcoded demo plans (replace with API call if you add GET /plans) ──
const plans = [
  { id: 1, name: 'Basic', cost: 9, features: ['1 user', '5 GB storage', 'Email support'] },
  { id: 2, name: 'Pro', cost: 29, features: ['5 users', '50 GB storage', 'Priority support'] },
  { id: 3, name: 'Enterprise', cost: 99, features: ['Unlimited users', '1 TB storage', '24/7 support'] },
]

const steps = ['Stripe Setup', 'Add Card', 'Subscribe', 'Done']

const currentStep = ref(0)
const stripeKey = ref(import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY || '')
const keyError = ref('')

const selectedPlanId = ref(1)
const subscription = ref(null)
const stripeError = ref('')
const cardFocused = ref(false)
const confirmingCard = ref(false)

let stripeInstance = null
let cardElement = null

// ── Step 0: init Stripe ──
async function initStripe() {
  keyError.value = ''
  if (!stripeKey.value.startsWith('pk_')) {
    keyError.value = 'Key must start with pk_test_ or pk_live_'
    return
  }

  stripeInstance = await loadStripe(stripeKey.value)
  currentStep.value = 1

  await nextTick()
  mountCardElement()
}

function mountCardElement() {
  const elements = stripeInstance.elements()
  cardElement = elements.create('card', {
    style: {
      base: {
        fontSize: '15px',
        color: '#1a1a2e',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
        '::placeholder': { color: '#aab' },
      },
      invalid: { color: '#e74c3c' },
    },
  })
  cardElement.mount('#card-element')
  cardElement.on('focus', () => (cardFocused.value = true))
  cardElement.on('blur', () => (cardFocused.value = false))
}

// ── Step 1: add payment method ──
const addCardMutation = useMutation({
  mutationFn: () => subscriptionApi.addPaymentMethod(),
  onSuccess: async (data) => {
    stripeError.value = ''
    confirmingCard.value = true
    try {
      const result = await stripeInstance.confirmCardSetup(data.clientSecret, {
        payment_method: { card: cardElement },
      })

      if (result.error) {
        stripeError.value = result.error.message
        return
      }
      // setup_intent.succeeded webhook fires → SetDefaultPaymentMethod on backend
      currentStep.value = 2
    } finally {
      confirmingCard.value = false
    }
  },
})

function handleAddCard() {
  stripeError.value = ''
  addCardMutation.mutate()
}

// ── Step 2: create subscription ──
const subscribeMutation = useMutation({
  mutationFn: () => subscriptionApi.createSubscription(selectedPlanId.value),
  onSuccess(data) {
    subscription.value = data
    currentStep.value = 3
  },
})

function handleSubscribe() {
  subscribeMutation.mutate()
}

// ── Reset ──
function resetFlow() {
  currentStep.value = 1
  subscription.value = null
  stripeError.value = ''
  addCardMutation.reset()
  subscribeMutation.reset()
  nextTick(mountCardElement)
}

// If env key is set, auto-skip step 0
onMounted(async () => {
  if (stripeKey.value.startsWith('pk_')) {
    await initStripe()
  }
})
</script>

<style scoped>
.center {
  display: flex;
  justify-content: center;
  padding-top: 20px;
}

.flow-card {
  background: #fff;
  border-radius: 16px;
  padding: 40px;
  width: 100%;
  max-width: 560px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
}

/* Stepper */
.stepper {
  display: flex;
  justify-content: space-between;
  margin-bottom: 36px;
  position: relative;
}
.stepper::before {
  content: '';
  position: absolute;
  top: 16px;
  left: 16px;
  right: 16px;
  height: 2px;
  background: #eee;
  z-index: 0;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  z-index: 1;
  flex: 1;
}

.step-circle {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #eee;
  color: #999;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  transition: all 0.3s;
}

.step.active .step-circle { background: #635bff; color: #fff; }
.step.done .step-circle { background: #2ecc71; color: #fff; }

.step-label {
  font-size: 11px;
  color: #aaa;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}
.step.active .step-label { color: #635bff; }
.step.done .step-label { color: #2ecc71; }

/* Section */
section h2 { font-size: 20px; margin-bottom: 8px; }
.desc { font-size: 14px; color: #888; margin-bottom: 24px; line-height: 1.5; }

/* Fields */
.field { margin-bottom: 18px; }

label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #555;
  margin-bottom: 6px;
}

input[type="text"],
input[type="number"] {
  width: 100%;
  padding: 11px 14px;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}
input:focus { border-color: #635bff; }

.hint { font-size: 12px; color: #aaa; margin-top: 4px; }

.stripe-wrapper {
  padding: 11px 14px;
  border: 1.5px solid #ddd;
  border-radius: 8px;
  min-height: 44px;
  transition: border-color 0.2s;
}
.stripe-wrapper.focused { border-color: #635bff; }

/* Plans */
.plans {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.plan-card {
  flex: 1;
  min-width: 140px;
  border: 2px solid #eee;
  border-radius: 10px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s;
}
.plan-card:hover { border-color: #c0bdff; }
.plan-card.selected { border-color: #635bff; background: #f5f4ff; }

.plan-name { font-size: 13px; font-weight: 700; margin-bottom: 6px; }
.plan-price { font-size: 22px; font-weight: 800; color: #635bff; margin-bottom: 10px; }
.plan-price span { font-size: 13px; font-weight: 400; color: #888; }

.plan-features { padding-left: 16px; }
.plan-features li { font-size: 12px; color: #666; margin-bottom: 4px; }

/* Buttons */
.btn-primary {
  width: 100%;
  padding: 13px;
  background: #635bff;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  margin-top: 8px;
  transition: background 0.2s;
}
.btn-primary:hover:not(:disabled) { background: #4e47d1; }
.btn-primary:disabled { background: #a0a0c0; cursor: not-allowed; }

.btn-secondary {
  width: 100%;
  padding: 13px;
  background: transparent;
  color: #635bff;
  border: 2px solid #635bff;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  margin-top: 8px;
  transition: all 0.2s;
}
.btn-secondary:hover { background: #f5f4ff; }

/* Alert */
.alert {
  padding: 12px 14px;
  border-radius: 8px;
  font-size: 14px;
  margin-bottom: 14px;
}
.alert.error { background: #fff0f0; color: #c0392b; }

/* Success */
.success-section { text-align: center; }

.success-icon {
  width: 64px;
  height: 64px;
  background: #2ecc71;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
  margin: 0 auto 20px;
}

.sub-details {
  background: #f8f8ff;
  border-radius: 10px;
  padding: 16px;
  margin: 20px 0;
  text-align: left;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  font-size: 14px;
  border-bottom: 1px solid #eee;
}
.detail-row:last-child { border-bottom: none; }
.detail-row span:first-child { color: #888; }
.detail-row span:last-child { font-weight: 600; font-size: 12px; word-break: break-all; }

.status-badge {
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}
.status-badge.active { background: #e6f9f0; color: #1a7a4a; }
.status-badge.incomplete { background: #fff8e1; color: #e67e22; }
.status-badge.canceled { background: #fff0f0; color: #c0392b; }
</style>
