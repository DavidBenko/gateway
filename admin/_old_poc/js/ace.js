App.AceEditorComponent = Ember.Component.extend({
  content: function(key, val) {
    if (!this.editor) {
      this.preset = val;
      return val;
    }
    if (arguments.length == 1) {
      return this.editor.getSession().getValue();
    }
    var cursor = this.editor.getCursorPosition();
    this.editor.getSession().setValue(val);
    this.editor.moveCursorToPosition(cursor);
    return val;
  }.property(),

  didInsertElement: function() {
    this.editor = window.ace.edit('editor');
    this.editor.setTheme('ace/theme/monokai');
    this.editor.getSession().setMode('ace/mode/javascript');

    var self = this;
    this.editor.on('change', function() {
      self.notifyPropertyChange('content');
    });

    if (this.preset) {
      this.set('content', this.preset);
      this.preset = null;
    }
  }
});
