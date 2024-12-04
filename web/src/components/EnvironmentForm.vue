<template>
  <v-form
    ref="form"
    lazy-validation
    v-model="formValid"
    v-if="item != null"
  >
    <!-- Environment Name -->
    <v-text-field
      v-model="item.name"
      :label="$t('environmentName')"
      :rules="[v => !!v || $t('name_required')]"
      required
      :disabled="formSaving"
      class="mb-4"
    ></v-text-field>

    <!-- Extra Variables Section -->
    <v-subheader class="px-0">
      Extra Variables
    </v-subheader>

    <v-text-field
      v-model="environmentVariables.env"
      label="Environment"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="environmentVariables.openshift_api"
      label="Openshift API"
      required
      :disabled="formSaving"
    ></v-text-field>

    <v-text-field
      v-model="environmentVariables.ac_backup_configuration.namespace"
      label="Namespace"
      required
      :disabled="formSaving"
    ></v-text-field>

    <!-- Restic Configuration Section -->
    <v-subheader class="px-0 mt-4">
      Restic Configuration
    </v-subheader>

    <v-select
      :items="resticOptions"
      item-text="name"
      item-value="id"
      v-model="selectedRestic"
      label="Select Restic Configuration"
      required
      :disabled="formSaving"
      @change="onResticSelected"
    ></v-select>

    <!-- Backup and Restore Mode Switch -->
    <v-radio-group v-model="mode" row>
      <v-radio label="Backup" value="backup"></v-radio>
      <v-radio label="Restore" value="restore"></v-radio>
    </v-radio-group>

    <!-- Backup Configuration -->
    <div v-if="mode === 'backup'">
      <!-- Backup Configuration Section -->
      <v-subheader class="px-0 mt-4">
        Backup Configuration
      </v-subheader>

      <v-text-field
        v-model.number="environmentVariables.ac_backup_configuration.timeout"
        label="Timeout"
        type="number"
        required
        :disabled="formSaving"
      ></v-text-field>

      <v-text-field
        v-model.number="environmentVariables.ac_backup_configuration.concurrent_backup_job_limit"
        label="Concurrent Backup Job Limit"
        type="number"
        required
        :disabled="formSaving"
      ></v-text-field>

      <!-- Additional Backup Options -->
      <v-checkbox
        v-model="environmentVariables.ac_backup_configuration.ac_backup_snapshot"
        label="Backup Snapshot"
        :disabled="formSaving"
      ></v-checkbox>

      <v-checkbox
        v-model="environmentVariables.ac_backup_configuration.ac_backup_dump"
        label="Backup Dump"
        :disabled="formSaving"
      ></v-checkbox>

      <!-- Email Configuration Section -->
      <v-subheader class="px-0 mt-4">
        Email Configuration
      </v-subheader>

      <v-text-field
        v-model="environmentVariables.ac_backup_configuration.email.mail_receiver"
        label="Mail Receiver"
        required
        :disabled="formSaving"
      ></v-text-field>

      <v-text-field
        v-model="environmentVariables.ac_backup_configuration.email.mail_from"
        label="Mail From"
        required
        :disabled="formSaving"
      ></v-text-field>

      <v-text-field
        v-model="environmentVariables.ac_backup_configuration.email.mail_subject"
        label="Mail Subject"
        required
        :disabled="formSaving"
      ></v-text-field>

      <v-text-field
        v-model="environmentVariables.ac_backup_configuration.email.mail_auth_user"
        label="Mail Auth User"
        required
        :disabled="formSaving"
      ></v-text-field>

      <v-text-field
        v-model="environmentVariables.ac_backup_configuration.email.mail_auth_password"
        label="Mail Auth Password"
        required
        :disabled="formSaving"
      ></v-text-field>

      <v-text-field
        v-model="environmentVariables.ac_backup_configuration.email.mail_server"
        label="Mail Server"
        required
        :disabled="formSaving"
      ></v-text-field>

      <!-- Backup Databases Section -->
      <v-subheader class="px-0 mt-4">
        Backup Databases
      </v-subheader>

      <v-data-table
        :items="environmentVariables.ac_backup_database"
        :headers="databaseHeaders"
        :items-per-page="-1"
        class="elevation-1"
        hide-default-footer
        no-data-text="No Databases"
      >
        <template v-slot:item="{ item }">
          <tr>
            <td class="pa-1">
              <v-text-field
                v-model="item.host"
                label="Host"
                solo-inverted
                flat
                hide-details
                class="v-text-field--solo--no-min-height"
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                v-model="item.user"
                label="User"
                solo-inverted
                flat
                hide-details
                class="v-text-field--solo--no-min-height"
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                v-model="item.password"
                label="Password"
                solo-inverted
                flat
                hide-details
                class="v-text-field--solo--no-min-height"
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                v-model="item.database"
                label="Database"
                solo-inverted
                flat
                hide-details
                class="v-text-field--solo--no-min-height"
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                v-model="item.type"
                label="Type"
                solo-inverted
                flat
                hide-details
                class="v-text-field--solo--no-min-height"
              ></v-text-field>
            </td>
            <!-- Cột Actions -->
            <td style="width: 38px;">
              <v-icon
                small
                class="pa-1"
                @click="removeDatabase(item)"
              >
                mdi-delete
              </v-icon>
            </td>
          </tr>
        </template>
      </v-data-table>

      <div class="text-right mt-2 mb-4">
        <v-btn
          color="primary"
          @click="addDatabase"
        >Add Database</v-btn>
      </div>
    </div>

    <!-- Restore Configuration -->
    <div v-else-if="mode === 'restore'">
      <!-- Restore PVCs Section -->
      <v-subheader class="px-0 mt-4">
        Restore PVCs
      </v-subheader>

      <v-data-table
        :items="restorePVCs"
        :headers="restorePVCHeaders"
        :items-per-page="-1"
        class="elevation-1"
        hide-default-footer
        no-data-text="No PVCs"
      >
        <template v-slot:item="{ item }">
          <tr>
            <td class="pa-1">
              <v-text-field
                v-model="item.src"
                label="Source PVC"
                solo-inverted
                flat
                hide-details
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                v-model="item.dest"
                label="Destination PVC"
                solo-inverted
                flat
                hide-details
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                v-model="item.version"
                label="Version"
                solo-inverted
                flat
                hide-details
              ></v-text-field>
            </td>
            <!-- Cột Actions -->
            <td style="width: 38px;">
              <v-icon
                small
                class="pa-1"
                @click="removeRestorePVC(item)"
              >
                mdi-delete
              </v-icon>
            </td>
          </tr>
        </template>
      </v-data-table>

      <div class="text-right mt-2 mb-4">
        <v-btn
          color="primary"
          @click="addRestorePVC"
        >
          Add Restore PVC
        </v-btn>
      </div>

      <!-- Restore Databases Section -->
      <v-subheader class="px-0 mt-4">
        Restore Databases
      </v-subheader>

      <v-data-table
        :items="restoreDBs"
        :headers="restoreDBHeaders"
        :items-per-page="-1"
        class="elevation-1"
        hide-default-footer
        no-data-text="No Databases"
      >
        <template v-slot:item="{ item }">
          <tr>
            <td class="pa-1">
              <v-text-field
                v-model="item.src"
                label="Source DB"
                solo-inverted
                flat
                hide-details
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                v-model="item.dest"
                label="Destination DB"
                solo-inverted
                flat
                hide-details
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                v-model="item.version"
                label="Version"
                solo-inverted
                flat
                hide-details
              ></v-text-field>
            </td>
            <td class="pa-1">
              <v-text-field
                v-model="item.password"
                label="Password"
                type="password"
                solo-inverted
                flat
                hide-details
              ></v-text-field>
            </td>
            <!-- Cột Actions -->
            <td style="width: 38px;">
              <v-icon
                small
                class="pa-1"
                @click="removeRestoreDB(item)"
              >
                mdi-delete
              </v-icon>
            </td>
          </tr>
        </template>
      </v-data-table>

      <div class="text-right mt-2 mb-4">
        <v-btn
          color="primary"
          @click="addRestoreDB"
        >
          Add Restore Database
        </v-btn>
      </div>
    </div>
  </v-form>
</template>

<script>
import ItemFormBase from '@/components/ItemFormBase';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';
import axios from 'axios';

export default {
  mixins: [ItemFormBase],

  data() {
    return {
      formValid: true,
      formError: '',
      formSaving: false,
      resticOptions: [],
      selectedRestic: null,
      mode: 'backup', // 'backup' hoặc 'restore'

      // Data variables for input fields
      environmentVariables: {
        env: '',
        disconnected_mode: false,
        openshift_api: '',
        openshift_client_file_tar: 'openshift-client-linux-4.14.0-0.okd-2023-12-01-225814.tar.gz',
        openshift_client_dir: '/tmp/semaphore/openshift-client',
        ac_backup_configuration: {
          timeout: null,
          k8up_annotation: 'ac/backup-schedule',
          ac_label: 'ac/backup',
          concurrent_backup_job_limit: null,
          namespace: '',
          backend: {
            repo_password: '',
            s3: {
              endpoint: '',
              bucket: '',
              username: '',
              password: '',
            },
          },
          ac_backup_snapshot: false,
          ac_backup_dump: false,
          ac_backup_dump_incremental: false,
          email: {
            mail_receiver: '',
            mail_from: '',
            mail_subject: '',
            mail_body: 'Backup report',
            mail_auth_user: '',
            mail_auth_password: '',
            mail_server: '',
          },
        },
        ac_backup_database: [], // Mảng trống để người dùng thêm database
        selectedResticId: null, // Lưu trữ ID của Restic đã chọn
      },

      // Restore data
      restorePVCs: [],
      restoreDBs: [],

      databaseHeaders: [
        { text: 'Host', value: 'host' },
        { text: 'User', value: 'user' },
        { text: 'Password', value: 'password' },
        { text: 'Database', value: 'database' },
        { text: 'Type', value: 'type' },
        { text: 'Actions', value: 'actions', sortable: false },
      ],

      restorePVCHeaders: [
        { text: 'Source PVC', value: 'src' },
        { text: 'Destination PVC', value: 'dest' },
        { text: 'Version', value: 'version' },
        { text: 'Actions', value: 'actions', sortable: false },
      ],

      restoreDBHeaders: [
        { text: 'Source DB', value: 'src' },
        { text: 'Destination DB', value: 'dest' },
        { text: 'Version', value: 'version' },
        { text: 'Password', value: 'password' },
        { text: 'Actions', value: 'actions', sortable: false },
      ],
    };
  },

  created() {
    this.fetchResticConfigs();
  },

  methods: {
    async fetchResticConfigs() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}/restic_configs`);
        this.resticOptions = response.data;

        // Nếu đang chỉnh sửa và đã chọn Restic config
        if (this.environmentVariables.selectedResticId) {
          this.selectedRestic = this.environmentVariables.selectedResticId;
          await this.onResticSelected();
        }
      } catch (error) {
        console.error('Error fetching Restic configs:', error);
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(error),
        });
      }
    },

    async fetchResticCredentials() {
      try {
        const response = await axios.get(`/api/project/${this.projectId}/credentials/${this.selectedRestic}`);

        const { username, password } = response.data;

        // Điền vào các trường S3
        const s3 = this.environmentVariables.ac_backup_configuration.backend.s3;
        s3.username = username;
        s3.password = password;
      } catch (error) {
        console.error('Error fetching Restic credentials:', error);

        // Xử lý trường hợp AccessKey không phải loại LoginPassword
        if (error.response && error.response.status === 400) {
          EventBus.$emit('i-snackbar', {
            color: 'error',
            text: 'Access Key type must be "login_password". Please check your Restic configuration.',
          });
        } else {
          EventBus.$emit('i-snackbar', {
            color: 'error',
            text: getErrorMessage(error),
          });
        }

        // Xóa giá trị của username và password nếu không lấy được
        const s3 = this.environmentVariables.ac_backup_configuration.backend.s3;
        s3.username = '';
        s3.password = '';
      }
    },

    async onResticSelected() {
      try {
        if (!this.selectedRestic) return;

        // Lấy chi tiết cấu hình Restic
        const res = await axios.get(`/api/project/${this.projectId}/restic_configs/${this.selectedRestic}`);
        const restic = res.data;

        // Điền các trường S3 với dữ liệu từ Restic
        const backend = this.environmentVariables.ac_backup_configuration.backend;
        backend.s3.endpoint = restic.url;
        backend.s3.bucket = restic.bucket;
        backend.repo_password = restic.restic_key;

        // Gọi API để lấy username và password
        await this.fetchResticCredentials();

        this.environmentVariables.selectedResticId = restic.id;
      } catch (error) {
        console.error('Lỗi khi lấy chi tiết Restic:', error);
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(error),
        });
      }
    },

    addDatabase() {
      this.environmentVariables.ac_backup_database.push({
        host: '',
        user: '',
        password: '',
        database: '',
        type: '',
      });
    },

    removeDatabase(item) {
      const index = this.environmentVariables.ac_backup_database.indexOf(item);
      if (index > -1) {
        this.environmentVariables.ac_backup_database.splice(index, 1);
      }
    },

    addRestorePVC() {
      this.restorePVCs.push({
        src: '',
        dest: '',
        version: '',
      });
    },

    removeRestorePVC(item) {
      const index = this.restorePVCs.indexOf(item);
      if (index > -1) {
        this.restorePVCs.splice(index, 1);
      }
    },

    addRestoreDB() {
      this.restoreDBs.push({
        src: '',
        dest: '',
        version: '',
        password: '',
      });
    },

    removeRestoreDB(item) {
      const index = this.restoreDBs.indexOf(item);
      if (index > -1) {
        this.restoreDBs.splice(index, 1);
      }
    },

    beforeSave() {
      // Save the selected Restic ID
      this.environmentVariables.selectedResticId = this.selectedRestic;

      // Thêm dữ liệu restore vào environmentVariables
      this.environmentVariables.ac_restore_pvc = this.restorePVCs;
      this.environmentVariables.ac_restore_db = this.restoreDBs;

      // Assemble the JSON from the input fields
      this.item.json = JSON.stringify(this.environmentVariables);
    },

    afterLoadData() {
      // Parse the JSON and populate the input fields
      try {
        const data = JSON.parse(this.item.json || '{}');
        this.environmentVariables = { ...this.environmentVariables, ...data };

        // Tải dữ liệu restore
        this.restorePVCs = data.ac_restore_pvc || [];
        this.restoreDBs = data.ac_restore_db || [];

        // Tải dữ liệu database
        this.environmentVariables.ac_backup_database = data.ac_backup_database || [];

        // Set the selected Restic ID and populate the S3 fields
        if (this.environmentVariables.selectedResticId) {
          this.selectedRestic = this.environmentVariables.selectedResticId;
          this.onResticSelected();
        }
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },

    saveForm() {
      if (this.$refs.form.validate()) {
        this.formSaving = true;
        this.beforeSave();
        this.saveItem()
          .then(() => {
            this.formSaving = false;
            this.$emit('save');
          })
          .catch((err) => {
            this.formSaving = false;
            this.formError = getErrorMessage(err);
          });
      }
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/environment`;
    },

    getSingleItemUrl() {
      return `/api/project/${this.projectId}/environment/${this.itemId}`;
    },
  },
};
</script>
