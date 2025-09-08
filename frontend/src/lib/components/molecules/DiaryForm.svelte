<script lang="ts">
import { _ } from "svelte-i18n";
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import Alert from "../atoms/Alert.svelte";
import Button from "../atoms/Button.svelte";
import FormField from "./FormField.svelte";

export let title: string;
export let content = "";
export let date = "";
export let error: string | undefined = undefined;
export let showDeleteButton = false;
export let onCancel: (() => void) | null = null;
export let onDelete: (() => void) | null = null;

let isSubmitting = false;

function _handleCancel() {
	if (onCancel) {
		onCancel();
	} else {
		goto("/");
	}
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">{title}</h1>
		<Button
			variant="secondary"
			size="md"
			on:click={_handleCancel}
		>
			{$_('diary.back')}
		</Button>
	</div>

	<div class="bg-white dark:bg-gray-800 shadow dark:shadow-gray-900/20 rounded-lg p-6">
		<form method="POST" use:enhance={() => {
			isSubmitting = true;
			return async ({ result }) => {
				isSubmitting = false;
			};
		}}>
			{#if error}
				<Alert type="error">
					{error}
				</Alert>
			{/if}

			<FormField
				type="input"
				inputType="date"
				label={$_('create.date')}
				id="date"
				name="date"
				required
				bind:value={date}
			/>

			<FormField
				type="textarea"
				label={$_('create.content')}
				id="content"
				name="content"
				placeholder={$_('diary.placeholder')}
				required
				rows={12}
				bind:value={content}
			/>

			<div class="flex {showDeleteButton ? 'justify-between' : 'justify-end'}">
				{#if showDeleteButton && onDelete}
					<Button
						type="button"
						variant="danger"
						size="md"
						on:click={onDelete}
					>
						{$_('diary.delete')}
					</Button>
				{/if}

				<div class="flex space-x-4">
					<Button
						type="button"
						variant="secondary"
						size="md"
						on:click={_handleCancel}
					>
						{$_('diary.cancel')}
					</Button>
					<Button
						type="submit"
						variant="primary"
						size="md"
						disabled={isSubmitting}
					>
						{#if isSubmitting}
							<div class="flex items-center">
								<svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								{$_('diary.saving')}
							</div>
						{:else}
							{$_('diary.save')}
						{/if}
					</Button>
				</div>
			</div>
		</form>
	</div>
</div>