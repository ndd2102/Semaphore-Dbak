<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null && keys != null">
    <EditDialog
      v-model="editDialog"
      :save-button-text="itemId === 'new' ? $t('create') : $t('save')"
      :title="`${itemId === 'new' ? $t('nnew') : $t('edit')} restic`"
      @save="loadItems()"
      :max-width="450"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <ResticForm
          :project-id="projectId"
          :item-id="itemId"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <ObjectRefsDialog
      object-title="restic_configs"
      :object-refs="itemRefs"
      :project-id="projectId"
      v-model="itemRefsDialog"
    />

    <YesNoDialog
      :title="'Delete snapshot'"
      :text="'Do you really want to delete this snapshot'"
      v-model="deleteItemDialog"
      @yes="deleteSnapshot"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>Snapshots</v-toolbar-title>
      <v-spacer></v-spacer>
    </v-toolbar>

    <v-menu
      v-model="menu"
      :close-on-content-click="false"
      offset-y
    >
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          color="primary"
          dark
          v-bind="attrs"
          v-on="on"
        >
          {{ selectedOption ? selectedOption.name : 'Select Restic' }}
        </v-btn>
      </template>

      <v-list>
        <v-list-item
          v-for="(option, index) in options"
          :key="index"
          @click="selectOption(option)"
        >
          <v-list-item-title>{{ option.name }}</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>

    <v-data-table
      v-if="selectedOption && selectedOption.id"
      :headers="headers"
      :items="dataSnapshots"
      hide-default-footer
      class="mt-4"
      :items-per-page="Number.MAX_VALUE"
    >
      <template v-slot:item.hostname="{ item }">
        {{ item.hostname }}
      </template>

      <template v-slot:item.paths="{ item }">
        {{ item.paths }}
      </template>

      <template v-slot:item.short_id="{ item }">
        {{ item.short_id }}
      </template>

      <template v-slot:item.size="{ item }">
        {{ item.size }}
      </template>

      <template v-slot:item.time="{ item }">
        {{ item.time }}
      </template>

      <template v-slot:item.actions="{ item }">
        <div style="white-space: nowrap">
          <v-btn
            icon
            class="mr-1"
            @click="askDeleteSnapshots(item.id)"
          >
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </div>
      </template>
    </v-data-table>
  </div>

</template>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import ResticForm from '@/components/ResticForm.vue';
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  mixins: [ItemListPageBase],
  components: { ResticForm },
  data() {
    return {
      keys: null,
      options: [],
      menu: false,
      selectedOption: null,
      dataSnapshots: [],
      deleteItemDialog: false,
      snapshotIdToDelete: null,
    };
  },

  async created() {
    this.keys = (await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/keys`,
      responseType: 'json',
    })).data;
    await this.getItemsRestic();
  },

  watch: {
    selectedOption: {
      handler(newval) {
        if (newval) {
          this.getItemsSnapshots();
        }
      },
      deep: true,
      immediate: true,
    },
  },
  methods: {
    async getItemsRestic() {
      const response = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/restic_configs`,
        responseType: 'json',
      })).data;
      console.log('rtesss', response);
      this.options = response.map((item) => ({
        name: item.name,
        id: item.id,
      }));
      if (this.options.length > 0 && !this.selectedOption) {
        this.selectedOption = this.options[0];
      }
    },
    async getItemsSnapshots() {
      if (this.selectedOption.id) {
        try {
          const response = (await axios({
            method: 'get',
            url: `/api/project/${this.projectId}/snapshots/${this.selectedOption.id}`,
            responseType: 'json',
          })).data;
          this.dataSnapshots = response;
        } catch (error) {
          this.dataSnapshots = [];
          EventBus.$emit('i-snackbar', {
            color: 'error',
            text: getErrorMessage(error),
          });
        }
      }
    },
    selectOption(item) {
      this.selectedOption = item;
      this.menu = false;
    },
    getHeaders() {
      return [{
        text: 'hostname',
        value: 'hostname',
      },
      {
        text: 'paths',
        value: 'paths',
      },
      {
        text: 'short_id',
        value: 'short_id',
      },
      {
        text: 'size',
        value: 'size',
      },
      {
        text: 'time',
        value: 'time',
      },
      {
        text: this.$i18n.t('actions'),
        value: 'actions',
        sortable: false,
      }];
    },
    getItemsUrl() {
      return `/api/project/${this.projectId}/restic_configs`;
    },
    getSingleItemUrl() {
      return `/api/project/${this.projectId}/restic_configs/${this.itemId}`;
    },
    getEventName() {
      return 'i-restic_configs';
    },
    askDeleteSnapshots(itemId) {
      this.snapshotIdToDelete = itemId;
      this.deleteItemDialog = true;
    },
    async deleteSnapshot() {
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}/snapshots/${this.selectedOption.id}?snapshot_id=${this.snapshotIdToDelete}`,
          responseType: 'json',
        });
        this.getItemsSnapshots();
      } catch (e) {
        // Do nothing
      } finally {
        // Đóng hộp thoại và đặt lại ID
        this.deleteItemDialog = false;
        this.snapshotIdToDelete = null;
      }
    },
  },
};
</script>
