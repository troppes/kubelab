<script ssr="false">
	import { toast } from '@zerodevx/svelte-toast';

	export let token;
	export let deployments;
	export let scaleDeployment;
	export let getConnectionString;
	export let teacherView;

	deployments = { items: [] };

	const errorToast = (message) => {
		toast.push(message, {
			theme: {
				'--toastColor': 'mintcream',
				'--toastBackground': '#f27474',
				'--toastBarBackground': '#fa5555'
			}
		});
	};

	const successToast = (message) => {
		toast.push(message, {
			theme: {
				'--toastColor': 'mintcream',
				'--toastBackground': 'rgba(72,187,120,0.9)',
				'--toastBarBackground': '#2F855A'
			}
		});
	};

	const connectionHandler = async (e) => {
		console.log('penis');
		console.log(e.srcElement.dataset.student);
		try {
			let deploy = deployments.items.find(
				(d) =>
					d.metadata.name == e.srcElement.dataset.id &&
					d.metadata.labels.student == e.srcElement.dataset.student
			);
			let string = await getConnectionString(
				token,
				{ nameSpace: deploy.metadata.labels.student, isTeacher: teacherView },
				deploy.metadata.name
			);
			navigator.clipboard
				.writeText(string)
				.then(() => successToast('Copied!'))
				.catch((e) => errorToast(e));
		} catch (error) {
			console.log(error);
		}
	};

	const scaleHandler = async (e) => {
		try {
			let deploy = deployments.items.find(
				(d) =>
					d.metadata.name == e.srcElement.dataset.id &&
					d.metadata.labels.student == e.srcElement.dataset.student
			);
			await scaleDeployment(
				token,
				{ nameSpace: deploy.metadata.labels.student, isTeacher: teacherView },
				deploy.metadata.name
			);
		} catch (error) {
			deployments = error;
		}
	};
</script>

<div class="item">
	{#await deployments}
		<div>
			<p>Fetching Classrooms ...</p>
		</div>
	{:then deployments}
		<div>
			<div>
				<table>
					<thead>
						<tr>
							<th>Name</th>
							<th>Status</th>
							<th>Action</th>
							<th>Connection</th>
						</tr>
					</thead>
					<tbody>
						{#each deployments.items as deploy}
							<tr>
								<td>
									{teacherView ? deploy.metadata.labels.student : deploy.metadata.name}
								</td>
								<td>
									{deploy.spec.replicas == 1 ? 'On' : 'Off'}
								</td>
								<td>
									<button
										class="button"
										data-id={deploy.metadata.name}
										data-student={deploy.metadata.labels.student}
										on:click={scaleHandler}>{deploy.spec.replicas == 1 ? 'Stop' : 'Start'}</button
									>
								</td>
								<td>
									<button
										class="button"
										data-id={deploy.metadata.name}
										data-student={deploy.metadata.labels.student}
										disabled={!(deploy.status.availableReplicas == 1)}
										on:click={connectionHandler}>Connect</button
									>
								</td>
							</tr>
							<tr class="details" data-id={deploy.metadata.name}>
								<td colspan="4" />
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{:catch error}
		<div>
			<p style="color: red">Error loading deployments.</p>
			<p style="color: red">Error message: {error.body.message}</p>
		</div>
	{/await}
</div>

<style>
	.details {
		display: none;
	}
	table {
		border-spacing: 10px;
		border-collapse: separate;
		text-align: center;
	}
</style>
