App.EnvironmentsIndexRoute = Ember.Route.extend({
  beforeModel: function() {
    this.transitionTo('newEnvironment');
  }
});

App.Environment = DS.Model.extend({
  name: DS.attr(),
  values: DS.attr()
});

App.EnvironmentsRoute = Ember.Route.extend({
  model: function() {
    return this.store.find('environment');
  }
});

App.EnvironmentsController = Ember.ArrayController.extend({
  sortProperties: ['name'],
  sortAscending: true
})

App.EnvironmentRoute = Ember.Route.extend({
  model: function(params) {
    return this.store.find('environment', params.environment_id);
  }
})

App.EnvironmentController = Ember.ObjectController.extend({
  needs: ["admin"],

  actions: {
    save: function() {
      var self = this;
      this.model.save().then(function(value) {
        self.set('controllers.admin.successMessage', "Saved!");
        self.set('controllers.admin.errorMessage', null);
      }, function(reason) {
        self.set('controllers.admin.successMessage', null);
        self.set('controllers.admin.errorMessage', reason.responseText);
      });
    },
    delete: function() {
      if (confirm("Delete the environment '" + this.model.get('name') + "'?")) {
        this.model.destroyRecord();
        this.set('controllers.admin.successMessage', "Deleted!");
        this.set('controllers.admin.errorMessage', null);
        this.transitionToRoute('environments');
      }
    }
  },
  values: function(key, val) {
    if (arguments.length == 2) {
      var v = null;
      try {
        v = JSON.parse(val);
      } catch (e) {}
      if (v) this.model.set('values', v);
    }
    return JSON.stringify(this.model.get('values'), null, 4);
  }.property('values', 'model')
});

App.NewEnvironmentRoute = Ember.Route.extend({
  templateName: 'environment',
  model: function(params) {
    return this.store.createRecord('environment');
  }
})

App.NewEnvironmentController = Ember.ObjectController.extend({
  // This is almost entirely duplicated from EnvironmentController,
  // but specifying controllerName in my route wouldn't resolve the
  // 'save' action.

  needs: ["admin"],

  actions: {
    save: function() {
      var self = this;
      this.model.save().then(function(value) {
        self.set('controllers.admin.successMessage', "Created!");
        self.set('controllers.admin.errorMessage', null);
        self.transitionToRoute("environment", value.id)
      }, function(reason) {
        self.set('controllers.admin.successMessage', null);
        self.set('controllers.admin.errorMessage', reason.responseText);
      });
    }
  },
  values: function(key, val) {
    if (arguments.length == 2) {
      var v = null;
      try {
        v = JSON.parse(val);
      } catch (e) {}
      if (v) this.model.set('values', v);
    }
    return JSON.stringify(this.model.get('values'), null, 4);
  }.property('values', 'model')
});
