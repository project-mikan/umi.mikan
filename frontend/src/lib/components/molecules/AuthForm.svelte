<script lang="ts">
import { _ } from "svelte-i18n";
import { enhance } from "$app/forms";
import Alert from "../atoms/Alert.svelte";
import Button from "../atoms/Button.svelte";
import Link from "../atoms/Link.svelte";
import FormField from "./FormField.svelte";

export let title: string;
export let submitText: string;
export let loadingText: string;
export let linkText: string;
export let linkHref: string;
export let showNameField = false;
export let error: string | undefined = undefined;
export let isRateLimited = false;

let loading = false;
let email = "";
let password = "";
let name = "";
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
	<div class="max-w-md w-full space-y-8">
		<div>
			<h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
				{title}
			</h2>
		</div>
		<form 
			class="mt-8 space-y-6" 
			method="POST" 
			use:enhance={({ formElement, formData, action, cancel }) => {
				loading = true;
				return async ({ result, update }) => {
					loading = false;
					await update();
				};
			}}
		>
			<div class="space-y-4">
				{#if showNameField}
					<FormField
						type="input"
						inputType="text"
						label={$_('auth.register.name')}
						id="name"
						name="name"
						autocomplete="name"
						placeholder={$_('auth.register.name')}
						required
						srOnlyLabel
						bind:value={name}
					/>
				{/if}

				<FormField
					type="input"
					inputType="email"
					label={$_('auth.login.email')}
					id="email"
					name="email"
					autocomplete="email"
					placeholder={$_('auth.login.email')}
					required
					srOnlyLabel
					bind:value={email}
				/>

				<FormField
					type="input"
					inputType="password"
					label={$_('auth.login.password')}
					id="password"
					name="password"
					autocomplete={showNameField ? 'new-password' : 'current-password'}
					placeholder={$_('auth.login.password')}
					required
					srOnlyLabel
					bind:value={password}
				/>
			</div>

			{#if error}
				<Alert type={isRateLimited ? "warning" : "error"}>
					{error}
				</Alert>
			{/if}

			<div>
				<Button
					type="submit"
					variant="primary"
					size="md"
					disabled={loading || isRateLimited}
					class="group relative w-full flex justify-center"
				>
					{loading ? loadingText : submitText}
				</Button>
			</div>

			<div class="text-center">
				<Link href={linkHref} variant="primary">
					{linkText}
				</Link>
			</div>
		</form>
	</div>
</div>