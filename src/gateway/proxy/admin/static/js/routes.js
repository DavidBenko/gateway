// Gateway routes -- not Ember routes.

App.RoutesRoute = Ember.Route.extend({
  model: function() {
    return $.ajax('routes').then(function(data){
      return JSON.parse(data);
    });
  }
});

App.RoutesController = Ember.ObjectController.extend({
  needs: ['admin'],
  actions: {
    update: function() {
      var self = this;
      $.ajax({
        type: "PUT",
        url: "routes",
        data: JSON.stringify(this.model)
      }).done(function( msg ) {
        self.set('controllers.admin.successMessage', "Saved!");
        self.set('controllers.admin.errorMessage', null);
      }).fail(function( jqXHR, textStatus ) {
        self.set('controllers.admin.successMessage', null);
        self.set('controllers.admin.errorMessage',
          "Request failed: " + textStatus );
      });
    }
  }
});
