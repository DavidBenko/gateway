// Gateway routes -- not Ember routes.

App.RoutesRoute = Ember.Route.extend({
  model: function() {
    return $.ajax('routes').then(function(data){
      return JSON.parse(data);
    });
  }
});

App.RoutesController = Ember.ObjectController.extend({
  actions: {
    update: function() {
      $.ajax({
        type: "PUT",
        url: "routes",
        data: JSON.stringify(this.model)
      }).done(function( msg ) {
        alert( "Data Saved: " + msg );
      }).fail(function( jqXHR, textStatus ) {
        alert( "Request failed: " + textStatus );
      });
    }
  }
});
